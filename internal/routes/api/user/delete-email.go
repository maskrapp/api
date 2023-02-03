package user

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
)

func DeleteEmail(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]string)
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		email, ok := body["email"]
		if !ok {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		userID := c.Locals("user_id").(string)
		db := ctx.Instances().Gorm
		err = db.Delete(&models.Email{}, "email = ? AND user_id = ?", email, userID).Error
		if err != nil {
			if strings.Contains(err.Error(), "(SQLSTATE 23503)") {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "There are still masks connected to that email. Delete those first.",
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "Email deleted",
		})
	}
}
