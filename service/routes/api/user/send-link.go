package user

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

func SendLink(postgrest *postgrest.Client, mailer *mailer.Mailer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		val, ok := body["email"]
		if !ok {
			return c.Status(400).SendString("Invalid Body")
		}

		email := val.(string)
		user := c.Locals("user").(*models.User)

		emailModel := &models.EmailEntry{}
		emailData, _, err := postgrest.From("emails").Select("*", "", false).Eq("user_id", user.ID).Eq("email", email).Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		err = json.Unmarshal(emailData, &emailModel)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}

		verification := &models.EmailVerificationEntry{
			EmailId:          emailModel.Id,
			VerificationCode: uuid.New().String(),
			ExpiresAt:        time.Now().Add(30 * time.Minute).Unix(),
		}
		_, _, err = postgrest.From("email_verifications").Insert(verification, true, "", "", "").Eq("email_id", strconv.Itoa(emailModel.Id)).Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		fullName := user.UserMetadata["full_name"].(string)
		err = mailer.SendVerifyMail(email, strings.Split(fullName, " ")[0], verification.VerificationCode)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return nil
	}
}
