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
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func GoogleHandler(handler *jwt.JWTHandler, db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		type body struct {
			Code string `json:"code"`
		}
		values := &body{}
		err := c.BodyParser(&values)
		if err != nil {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		data, err := extractGoogleData(values.Code)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		provider := &models.Provider{
			ID:           data.Id,
			ProviderName: "google",
		}

		var user *models.User

		err = db.First(provider).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return c.SendStatus(500)
		}
		if err == gorm.ErrRecordNotFound {
			usr, err := createGoogleUser(db, data)
			user = usr
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			err = db.Create(&models.Provider{
				ID:           data.Id,
				ProviderName: "google",
				UserID:       user.ID,
			}).Error
			if err != nil {
				return c.SendStatus(500)
			}
		} else {
			usr := &models.User{
				ID: provider.UserID,
			}
			err := db.First(usr).Error
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			user = usr
		}
		pair, err := handler.CreatePair(user.ID, user.TokenVersion)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(pair)
	}
}

func createGoogleUser(db *gorm.DB, data *GoogleData) (*models.User, error) {
	uuid := uuid.New()
	user := &models.User{
		ID:            uuid.String(),
		Role:          0,
		Password:      nil,
		Name:          data.GivenName,
		Email:         data.Email,
		EmailVerified: data.VerifiedEmail,
	}
	err := db.Create(user).Error
	if err != nil {
		return nil, err
	}
	fmt.Println("user", user)
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
