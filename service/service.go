package service

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/service/routes"
	"github.com/maskrapp/common/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BackendService struct {
	fiber  *fiber.App
	mailer *mailer.Mailer
	db     *gorm.DB
	logger *logrus.Logger
}

func New(emailToken, templateToken, jwtSecret string, dbUser, dbPassword, dbHost, dbDatabase string, production bool) *BackendService {
	uri := fmt.Sprintf("postgres://%v:%v@%v/%v", dbUser, dbPassword, dbHost, dbDatabase)
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	service := &BackendService{
		mailer: mailer.New(emailToken, templateToken, production),
		fiber:  fiber.New(),
		db:     db,
		logger: logrus.New(),
	}

	routes.Setup(service.fiber, service.mailer, jwt.New(os.Getenv("SECRET_KEY"), time.Minute*5, time.Hour*24), db, service.logger)
	return service
}

func (b *BackendService) Start() {
	b.db.AutoMigrate(&models.User{}, &models.Email{}, &models.EmailVerification{}, &models.Mask{}, &models.Provider{})
	b.fiber.Listen(":80")
}

func (b *BackendService) Shutdown() {
	b.fiber.Shutdown()
}
