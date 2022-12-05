package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/routes"
	"github.com/maskrapp/common/models"
)

type Service struct {
	fiber *fiber.App
}

func New(ctx global.Context) (*Service, error) {
	err := ctx.Instances().Gorm.AutoMigrate(&models.User{}, &models.Email{}, &models.EmailVerification{}, &models.Mask{}, &models.Provider{}, &models.AccountVerification{}, &models.Domain{})

	if err != nil {
		return nil, err
	}

	svc := &Service{fiber.New()}
	routes.Setup(ctx, svc.fiber)
	return svc, nil
}

func (b *Service) Start() error {
	return b.fiber.Listen(":80")
}
func (b *Service) Shutdown() {
	b.fiber.Shutdown()
}
