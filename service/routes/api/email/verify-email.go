package email

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"github.com/supabase/postgrest-go"
)

func VerifyEmail(postgrest *postgrest.Client) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		code, ok := body["code"].(string)
		if !ok {
			return c.SendStatus(400)
		}
		verificationModel := &models.EmailVerificationEntry{}
		_, err = postgrest.From("email_verifications").Select("*", "", false).Eq("verification_code", code).Single().ExecuteTo(verificationModel)
		if err != nil {
			if strings.Contains(err.Error(), "(PGRST116)") {
				return c.Status(404).JSON(&models.APIResponse{
					Success: false,
					Message: "Incorrect code",
				})
			}
			return c.SendStatus(500)
		}
		if time.Now().Unix() > verificationModel.ExpiresAt {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Code has expired",
			})
		}
		_, _, err = postgrest.From("email_verifications").Delete("", "").Eq("id", strconv.Itoa(verificationModel.Id)).Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		values := make(map[string]bool)
		values["is_verified"] = true
		_, _, err = postgrest.From("emails").Update(values, "", "").Eq("id", strconv.Itoa(verificationModel.EmailId)).Single().Execute()
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return nil
	}
}
