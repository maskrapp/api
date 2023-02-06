package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func GoogleHandler(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		type body struct {
			Code string `json:"code"`
		}

		values := &body{}
		err := c.BodyParser(&values)
		if err != nil {
			return c.Status(fiber.ErrBadRequest.Code).JSON(
				&models.APIResponse{
					Success: false,
					Message: "Invalid body",
				},
			)
		}
		data, err := extractGoogleData(values.Code)
		if err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(
				&models.APIResponse{
					Success: false,
					Message: "Google token exchange error",
				},
			)
		}
		provider := &models.Provider{
			ID:           data.Id,
			ProviderName: "google",
		}
		var user *models.User
		db := ctx.Instances().Gorm
		err = db.First(provider).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			logrus.Error("Database error:", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		if err == gorm.ErrRecordNotFound {
			usr, err := createGoogleUser(ctx.Instances().Gorm, data)
			user = usr
			if err != nil {
				logrus.Error("Database error:", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
					Message: "Something went wrong!",
				})
			}
			err = db.Create(&models.Provider{
				ID:           data.Id,
				ProviderName: "google",
				UserID:       user.ID,
			}).Error
			if err != nil {
				logrus.Error("Database error:", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
					Message: "Something went wrong!",
				})
			}
			err = db.Create(&models.Email{
				UserID:     user.ID,
				IsPrimary:  true,
				IsVerified: true,
				Email:      data.Email,
			}).Error
			if err != nil {
				logrus.Error("Database error:", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
					Message: "Something went wrong!",
				})
			}
		} else {
			usr := &models.User{
				ID: provider.UserID,
			}
			err := db.First(usr).Error
			if err != nil {
				logrus.Error("Database error:", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
					Message: "Something went wrong!",
				})
			}
			user = usr
		}
		pair, err := ctx.Instances().JWT.CreatePair(user.ID, user.TokenVersion, "google")
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(pair)
	}
}

func createGoogleUser(db *gorm.DB, data *GoogleData) (*models.User, error) {
	uuid := uuid.New()
	user := &models.User{
		ID:    uuid.String(),
		Role:  0,
		Email: data.Email,
	}
	err := db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT"),
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_SECRET"),
	Endpoint:     google.Endpoint,
	Scopes:       []string{"profile", "email"},
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

type GoogleData struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	PictureURL    string `json:"picture"`
	Locale        string `json:"locale"`
}

func extractGoogleData(code string) (*GoogleData, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	googleData := &GoogleData{}
	err = json.Unmarshal(data, googleData)
	if err != nil {
		return nil, err
	}
	return googleData, nil
}
