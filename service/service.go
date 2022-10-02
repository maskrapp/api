package service

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	"github.com/maskrapp/backend/service/routes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BackendService struct {
	fiber  *fiber.App
	mailer *mailer.Mailer
	db     *gorm.DB
}

func New(emailToken, templateToken, jwtSecret string, production bool, postgresURI string) *BackendService {
	db, err := gorm.Open(postgres.Open(postgresURI), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	service := &BackendService{
		mailer: mailer.New(emailToken, templateToken, production),
		fiber:  fiber.New(),
		db:     db,
	}

	routes.Setup(service.fiber, service.mailer, jwt.New(os.Getenv("SECRET_KEY"), time.Minute*5, time.Hour*24), db)
	return service
}

func (b *BackendService) Start() {
	b.db.AutoMigrate(&models.User{}, &models.Email{}, &models.EmailVerification{}, &models.Mask{}, &models.Provider{})
	b.fiber.Listen(":80")
}

func (b *BackendService) Shutdown() {
	b.fiber.Shutdown()
}
