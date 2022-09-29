package user

import (
	"encoding/json"
	"strings"

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
		val, ok := body["email"]
		if !ok {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		email := val.(string)
		user := c.Locals("user").(*models.User)
		data, _, err := postgrest.From("emails").Delete("", "").Eq("user_id", user.ID).Eq("email", email).Eq("is_primary", "false").ExecuteString()
		if err != nil || len(data) < 3 {
			if err != nil && strings.Contains(err.Error(), "(23503)") {
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
