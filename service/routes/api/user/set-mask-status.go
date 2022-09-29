package user

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

type setMaskBody struct {
	Mask  string `json:"mask"`
	Value bool   `json:"value"`
}

func SetMaskStatus(postgrest *postgrest.Client) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		b := &setMaskBody{}
		err := json.Unmarshal(c.Body(), &b)
		if err != nil {
			return c.SendStatus(500)
		}
		if b.Mask == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		user := c.Locals("user").(*models.User)
		values := map[string]string{
			"enabled": strconv.FormatBool(b.Value),
		}
		result, _, err := postgrest.From("masks").Update(values, "", "").Eq("user_id", user.ID).Eq("mask", b.Mask).Single().ExecuteString()
		if len(result) == 0 || err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
		})
	}
}
