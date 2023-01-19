package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
)

func AccountDetails(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := &models.User{}
		userId := c.Locals("user_id").(string)
		err := ctx.Instances().Gorm.Find(user, "id = ?", userId).Error
		if err != nil {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		details := make(map[string]interface{})
		details["email"] = user.Email

		return c.JSON(details)
	}
}
