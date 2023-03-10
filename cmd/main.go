package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/maskrapp/api/internal/config"
	"github.com/maskrapp/api/internal/domains"
	"github.com/maskrapp/api/internal/global"
	grpc_impl "github.com/maskrapp/api/internal/grpc"
	"github.com/maskrapp/api/internal/healthcheck"
	"github.com/maskrapp/api/internal/jwt"
	"github.com/maskrapp/api/internal/mailer"
	"github.com/maskrapp/api/internal/models"
	main_api "github.com/maskrapp/api/internal/pb/main_api/v1"
	"github.com/maskrapp/api/internal/ratelimit"
	"github.com/maskrapp/api/internal/recaptcha"
	"github.com/maskrapp/api/internal/routes"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.New()

	level, err := logrus.ParseLevel(cfg.Logger.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	logrus.SetLevel(level)

	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host,
		Password: cfg.Redis.Password,
	})
	status := redis.Conn().Ping(context.Background())
	err = status.Err()

	if err != nil {
		logrus.Panic(err)
	}

	uri := fmt.Sprintf("postgres://%v:%v@%v/%v", cfg.Database.Username, cfg.Database.Password, cfg.Database.Hostname, cfg.Database.Database)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		logrus.Panic(err)
	}

	domainService := domains.New(db, time.Minute*2)
	domainService.Start()

	instances := &global.Instances{
		Gorm:        db,
		Redis:       redis,
		RateLimiter: ratelimit.New(redis, 50, map[string]int{}),
		Recaptcha:   recaptcha.New(cfg.Recaptcha.Secret),
		JWT:         jwt.New(cfg.JWT.Secret, 5*time.Minute, 24*time.Hour),
		Mailer:      mailer.New(cfg),
		Domains:     domainService,
	}

	gCtx, cancel := global.WithCancel(global.NewContext(context.Background(), instances, cfg))

	defer cancel()

	err = instances.Gorm.AutoMigrate(&models.User{}, models.Email{}, models.EmailVerification{}, models.Mask{}, models.Provider{}, models.AccountVerification{}, models.Domain{}, models.PasswordResetVerification{})
	if err != nil {
		logrus.Panic(err)
	}

	fiber := fiber.New()
	routes.Setup(gCtx, fiber)

	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionAge: 1 * time.Minute,
	}))
	main_api.RegisterMainAPIServiceServer(grpcServer, grpc_impl.NewMainAPIService(gCtx))

	address := fmt.Sprintf(":%v", cfg.GRPC.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		logrus.Panic(err)
	}

	logrus.Infof("Service Info: %v", grpcServer.GetServiceInfo())

	healthCheckSrv := healthcheck.New(gCtx)

	logrus.Infof("listening for GRPC requests on port: %v", cfg.GRPC.Port)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go grpcServer.Serve(ln)
	go fiber.Listen(":80")
	go healthCheckSrv.ListenAndServe()

	<-shutdownChan
	logrus.Info("gracefully shutting down...")
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		healthCheckSrv.Shutdown(c)
	}()
	go func() {
		defer wg.Done()
		fiber.ShutdownWithTimeout(10 * time.Second)
	}()
	go func() {
		defer wg.Done()
		grpcServer.GracefulStop()
	}()
	wg.Wait()
}
