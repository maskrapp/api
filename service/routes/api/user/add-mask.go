package user

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

type addMaskBody struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Domain string `json:"domain"`
}

func AddMask(postgrest *postgrest.Client) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		b := &addMaskBody{}
		err := json.Unmarshal(c.Body(), &b)
		if err != nil {
			return c.SendStatus(500)
		}
		if b.Email == "" || b.Domain == "" || b.Name == "" {
			return c.Status(400).SendString("Invalid Body")
		}
		fullEmail := b.Name + "@" + b.Domain
		match := emailRegex.MatchString(fullEmail)
		if !match {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid masked email address",
			})
		}
		if b.Domain != "relay.maskr.app" {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Domain not found",
			})
		}
		result, _, err := postgrest.From("masks").Select("*", "", false).Eq("mask", fullEmail).Single().ExecuteString()
		if err != nil && !strings.Contains(err.Error(), "(PGRST116)") {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		if len(result) > 0 {
			return c.Status(409).JSON(&models.APIResponse{
				Success: false,
				Message: "Mask already in use",
			})
		}
		emailEntry := &models.Email{}
		user := c.Locals("user").(*models.User)
		_, err = postgrest.From("emails").Select("*", "", false).Eq("user_id", user.ID).Eq("email", b.Email).Single().ExecuteTo(emailEntry)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})
		}
		if !emailEntry.IsVerified {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Email is not verified",
			})
		}
		maskEntry := &models.Mask{
			UserID:    user.ID,
			Enabled:   true,
			ForwardTo: emailEntry.Id,
			Mask:      fullEmail,
		}
		_, _, err = postgrest.From("masks").Insert(maskEntry, false, "", "", "").Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
			Message: "Created mask",
		})
	}
}
