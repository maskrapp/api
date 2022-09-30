package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
)

func AuthMiddleware(supabaseToken, baseURL string) func(*fiber.Ctx) error {
	var endpoint = baseURL + "/auth/v1/user"
	return func(c *fiber.Ctx) error {
		auth := c.GetReqHeaders()["Authorization"]
		if auth == "" {
			return c.SendStatus(401)
		}
		client := http.DefaultClient
		request, _ := http.NewRequest("GET", endpoint, nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
		request.Header.Set("apiKey", supabaseToken)
		response, err := client.Do(request)
		if err != nil || response.StatusCode != 200 {
			return c.SendStatus(500)
		}
		usr := &models.User{}
		err = json.NewDecoder(response.Body).Decode(&usr)
		if err != nil {
			return c.SendStatus(500)
		}
		if usr.Role != 0 {
			return c.SendStatus(401)
		}
		c.Locals("user", usr)
		return c.Next()
	}
}
