package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/service/middleware"
	apiauth "github.com/maskrapp/backend/service/routes/api/auth"
	"github.com/maskrapp/backend/service/routes/api/email"
	"github.com/maskrapp/backend/service/routes/api/user"
	"github.com/maskrapp/backend/service/routes/auth"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, mailer *mailer.Mailer, jwtHandler *jwt.JWTHandler, gorm *gorm.DB) {
	app.Use(cors.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})

	authGroup := app.Group("/auth")
	authGroup.Post("/google", auth.GoogleHandler(jwtHandler, gorm))

	apiGroup := app.Group("/api")

	apiUserGroup := apiGroup.Group("/user")
	apiUserGroup.Use(middleware.AuthMiddleware(jwtHandler))

	apiUserGroup.Post("/emails", user.Emails(gorm))
	apiUserGroup.Post("/add-email", user.AddEmail(gorm, mailer))
	apiUserGroup.Delete("/delete-email", user.DeleteEmail(gorm))

	apiUserGroup.Post("/masks", user.Masks(gorm))
	apiUserGroup.Post("add-mask", user.AddMask(gorm))
	apiUserGroup.Delete("delete-mask", user.DeleteMask(gorm))
	apiUserGroup.Put("set-mask-status", user.SetMaskStatus(gorm))

	apiUserGroup.Post("/send-link", user.SendLink(gorm, mailer))

	apiEmailGroup := apiGroup.Group("/email")
	apiEmailGroup.Post("/verify", email.VerifyEmail(gorm))

	apiAuthGroup := apiGroup.Group("/auth")
	apiAuthGroup.Post("/refresh", apiauth.RefreshToken(jwtHandler))

}
