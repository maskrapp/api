package user

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

func DeleteEmail(postgrest *postgrest.Client) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		email := body["email"].(string)
		if email == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		user := c.Locals("user").(*models.User)
		data, _, err := postgrest.From("emails").Delete("", "").Eq("user_id", user.ID).Eq("email", email).Eq("is_primary", "false").ExecuteString()
		if err != nil || len(data) < 3 {
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
