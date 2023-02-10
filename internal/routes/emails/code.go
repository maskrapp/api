package emails

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Verify is used for verifying unverified email addresses.
// This endpoint is accessible at: POST /emails/{email}/verify
func Verify(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Code string `json:"code"`
		}
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		if body.Code == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		email := c.Params("email")
		if email == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid parameters",
			})
		}

		userID := c.Locals("user_id").(string)
		emailModel := &models.Email{}
		db := ctx.Instances().Gorm
		err = db.Find(emailModel, "user_id = ? AND email = ?", userID, email).Error
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid email",
			})
		}
		if emailModel.IsVerified {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That email is already verified",
			})
		}

		verificationModel := &models.EmailVerification{}
		err = db.Find(verificationModel, "email_id = ?", emailModel.Id).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// User has to request a verification code first.
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		if body.Code != verificationModel.VerificationCode {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid code",
			})
		}

		if time.Now().Unix() > verificationModel.ExpiresAt {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Code has expired",
			})
		}

		err = db.Delete(&models.EmailVerification{}, "id", verificationModel.Id).Error
		if err != nil {
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		values := make(map[string]interface{})
		values["is_verified"] = true
		err = db.Model(&models.Email{}).Where("id = ?", verificationModel.EmailID).Updates(values).Error
		if err != nil {
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}

// RequestCode is used for verifying unverified email addresses.
// It is the first step in this process. On success, the server will send a verification code to to the user's desired email. This code can then be redeemed at the Verify endpoint.
// This endpoint is accessible at: POST /emails/{email}/create-code
func RequestCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		email := c.Params("email")
		if email == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid parameters",
			})
		}

		userID := c.Locals("user_id").(string)

		emailRecord := &models.Email{}
		db := ctx.Instances().Gorm
		err := db.Find(emailRecord, "email = ? AND user_id = ?", email, userID).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Email not found",
				})
			}
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		if emailRecord.IsVerified {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That email is already verified",
			})
		}

		verification := &models.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: utils.GenerateCode(5),
			ExpiresAt:        time.Now().Add(5 * time.Minute).Unix(),
		}
		if db.Model(&verification).Where("email_id = ?", emailRecord.Id).Updates(&verification).RowsAffected == 0 {
			err = db.Create(&verification).Error
		}
		if err != nil {
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		err = ctx.Instances().Mailer.SendVerifyMail(email, verification.VerificationCode)
		if err != nil {
			logrus.Errorf("mailer error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "A verification code has been sent to your email",
		})
	}
}
