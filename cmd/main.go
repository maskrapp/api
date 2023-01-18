package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/maskrapp/backend/internal/config"
	"github.com/maskrapp/backend/internal/domains"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/jwt"
	"github.com/maskrapp/backend/internal/mailer"
	"github.com/maskrapp/backend/internal/ratelimit"
	"github.com/maskrapp/backend/internal/recaptcha"
	"github.com/maskrapp/backend/internal/routes"
	"github.com/maskrapp/backend/internal/utils"
	"github.com/maskrapp/common/models"
	"github.com/sirupsen/logrus"
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

	httpClient := utils.CreateCustomHttpClient()

	instances := &global.Instances{
		Gorm:        db,
		Redis:       redis,
		RateLimiter: ratelimit.New(redis, 50, map[string]int{}),
		Recaptcha:   recaptcha.New(httpClient, cfg.Recaptcha.Secret),
		JWT:         jwt.New(cfg.JWT.Secret, 5*time.Minute, 24*time.Hour),
		Mailer:      mailer.New(httpClient, cfg.ZeptoMail.EmailToken, cfg.ZeptoMail.TemplateKey, cfg.Production),
		Domains:     domains.New(db, 10*time.Minute),
	}

	gCtx, cancel := global.WithCancel(global.NewContext(context.Background(), instances, cfg))
	defer cancel()

	err = instances.Gorm.AutoMigrate(&models.User{}, &models.Email{}, &models.EmailVerification{}, &models.Mask{}, &models.Provider{}, &models.AccountVerification{}, &models.Domain{})
	if err != nil {
		logrus.Panic(err)
	}

	fiber := fiber.New()
	routes.Setup(gCtx, fiber)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go fiber.Listen(":80")

	<-shutdownChan
	logrus.Info("gracefully shutting down...")
	fiber.ShutdownWithTimeout(10 * time.Second)
}
