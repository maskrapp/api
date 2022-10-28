package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/models"
	"github.com/maskrapp/backend/utils"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

type createAccountBody struct {
	Email    string
	Code     string
	Password string
}

func CreateAccount(db *gorm.DB, jwtHandler *jwt.JWTHandler) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		body := &createAccountBody{}

		err := c.BodyParser(body)

		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if body.Email == "" || body.Code == "" || body.Password == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		verificationRecord := &dbmodels.AccountVerification{}
		err = db.First(verificationRecord, "email = ? AND verification_code = ? ", body.Email, body.Code).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Invalid code",
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		err = db.Delete(&dbmodels.AccountVerification{}, "email = ?", body.Email).Error

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		hashedPassword, err := utils.HashPassword(body.Password)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		user := &dbmodels.User{ID: uuid.NewString(), Name: "", Role: 0, Password: &hashedPassword, Email: body.Email}
		err = db.Create(user).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		provider := &dbmodels.Provider{ID: uuid.NewString(), ProviderName: "email", UserID: user.ID}

		err = db.Create(provider).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		err = db.Create(&dbmodels.Email{
			UserID:     user.ID,
			IsPrimary:  true,
			IsVerified: true,
			Email:      body.Email,
		}).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		pair, err := jwtHandler.CreatePair(user.ID, user.TokenVersion)

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		return c.JSON(pair)

	}
}
