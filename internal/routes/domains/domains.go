package domains

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
)

// Get is used for retrieving the available domains
// This endpoint is accessible at GET /domains
func Get(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		availableDomains := ctx.Instances().Domains.Values()
		return c.JSON(availableDomains)
	}
}
