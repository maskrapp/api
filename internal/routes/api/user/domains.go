package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
)

func Domains(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		availableDomains := ctx.Instances().Domains.Values()
		return c.JSON(availableDomains)
	}
}
