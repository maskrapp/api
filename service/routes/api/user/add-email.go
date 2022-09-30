package user

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

func AddEmail(postgrest *postgrest.Client, mailer *mailer.Mailer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		val, ok := body["email"]
		if !ok {
			return c.Status(400).SendString("Invalid Body")
		}
		email := val.(string)

		user := c.Locals("user").(*models.User)
		//maybe just upsert this?
		emailData, _, err := postgrest.From("emails").Select("*", "", false).Filter("user_id", "eq", user.ID).Filter("email", "eq", email).Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		emails := make([]struct{}, 0)
		err = json.Unmarshal(emailData, &emails)
		if err != nil {
			return c.SendStatus(500)
		}
		if len(emails) > 0 {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That email is already registered to your account",
			})
		}
		emailEntry := &models.Email{
			UserID:     user.ID,
			IsPrimary:  false,
			IsVerified: false,
			Email:      email,
		}
		data, _, err := postgrest.From("emails").Insert(emailEntry, false, "", "", "").Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		result := &models.Email{}
		err = json.Unmarshal(data, result)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}

		emailVerification := &models.EmailVerification{
			EmailID:          result.Id,
			VerificationCode: uuid.New().String(),
			ExpiresAt:        time.Now().Add(30 * time.Minute).Unix(),
		}
		_, _, err = postgrest.From("email_verifications").Insert(emailVerification, true, "", "", "").Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		fullName := "unknown"
		err = mailer.SendVerifyMail(email, strings.Split(fullName, " ")[0], emailVerification.VerificationCode)
		if err != nil {
			return err
		}
		return nil
	}
}
