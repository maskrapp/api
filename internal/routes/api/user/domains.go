package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	dbmodels "github.com/maskrapp/common/models"
)

func Domains(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var availableDomains []dbmodels.Domain
		err := ctx.Instances().Gorm.Find(&availableDomains).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(availableDomains)
	}
}
