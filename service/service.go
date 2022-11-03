package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/ratelimit"
	"github.com/maskrapp/backend/service/routes"
	"github.com/maskrapp/common/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BackendService struct {
	fiber       *fiber.App
	mailer      *mailer.Mailer
	db          *gorm.DB
	logger      *logrus.Logger
	redis       *redis.Client
	ratelimiter *ratelimit.RateLimiter
}

func New(emailToken, templateToken, jwtSecret, dbUser, dbPassword, dbHost, dbDatabase, redisHost, redisPassword string, production bool, globalRPI int, customRoutes map[string]int) *BackendService {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
	})

	//TODO: skip step if err
	status := redisClient.Conn().Ping(context.TODO())
	err := status.Err()
	if err != nil {
		panic(err)
	}

	logger := logrus.New()

	rateLimiter := ratelimit.New(redisClient, logger, globalRPI, customRoutes)

	uri := fmt.Sprintf("postgres://%v:%v@%v/%v", dbUser, dbPassword, dbHost, dbDatabase)
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		panic(err)

	}
	service := &BackendService{
		mailer:      mailer.New(emailToken, templateToken, production),
		fiber:       fiber.New(),
		db:          db,
		logger:      logger,
		redis:       redisClient,
		ratelimiter: rateLimiter,
	}

	routes.Setup(service.fiber, service.mailer, jwt.New(os.Getenv("SECRET_KEY"), time.Minute*5, time.Hour*24), db, service.logger, rateLimiter)
	return service
}

func (b *BackendService) Start() {
	err := b.db.AutoMigrate(&models.User{}, &models.Email{}, &models.EmailVerification{}, &models.Mask{}, &models.Provider{}, &models.AccountVerification{})
	if err != nil {
		panic(err)
	}
	b.fiber.Listen(":80")
}
func (b *BackendService) Shutdown() {
	b.fiber.Shutdown()
}
