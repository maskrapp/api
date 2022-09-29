package user

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

func DeleteMask(postgrest *postgrest.Client) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		val, ok := body["mask"]
		if !ok {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		mask := val.(string)
		user := c.Locals("user").(*models.User)
		result, _, err := postgrest.From("masks").Delete("", "").Eq("user_id", user.ID).Eq("mask", mask).ExecuteString()
		if err != nil || len(result) < 3 {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "That mask does not exist",
			})
		}
		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
			Message: "Succesfully deleted mask",
		})
	}
}
