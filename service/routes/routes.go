package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/ratelimit"
	"github.com/maskrapp/backend/recaptcha"
	"github.com/maskrapp/backend/service/middleware"
	apiauth "github.com/maskrapp/backend/service/routes/api/auth"
	"github.com/maskrapp/backend/service/routes/api/user"
	"github.com/maskrapp/backend/service/routes/auth"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, mailer *mailer.Mailer, jwtHandler *jwt.JWTHandler, gorm *gorm.DB, ratelimiter *ratelimit.RateLimiter, recaptcha *recaptcha.Recaptcha) {
	app.Use(cors.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})
	app.Get("/headers", func(c *fiber.Ctx) error {
		return c.JSON(c.GetReqHeaders())
	})
	authGroup := app.Group("/auth")
	authGroup.Post("/google", auth.GoogleHandler(jwtHandler, gorm))

	authGroup.Post("create-account-code", auth.CreateAccountCode(gorm, jwtHandler, mailer, recaptcha))
	authGroup.Post("verify-account-code", auth.VerifyAccountCode(gorm, recaptcha))
	authGroup.Post("resend-account-code", auth.ResendAccountCode(gorm, mailer, recaptcha))
	authGroup.Post("create-account", auth.CreateAccount(gorm, jwtHandler, recaptcha))
	authGroup.Post("email-login", auth.EmailLogin(gorm, jwtHandler, recaptcha))

	apiGroup := app.Group("/api")

	apiUserGroup := apiGroup.Group("/user")
	apiUserGroup.Use(middleware.AuthMiddleware(jwtHandler))
	apiUserGroup.Use(middleware.UserRateLimit(ratelimiter))

	apiUserGroup.Post("/emails", user.Emails(gorm))
	apiUserGroup.Post("/add-email", user.AddEmail(gorm, mailer))
	apiUserGroup.Delete("/delete-email", user.DeleteEmail(gorm))

	apiUserGroup.Post("/masks", user.Masks(gorm))
	apiUserGroup.Post("add-mask", user.AddMask(gorm))
	apiUserGroup.Delete("delete-mask", user.DeleteMask(gorm))
	apiUserGroup.Put("set-mask-status", user.SetMaskStatus(gorm))

	apiUserGroup.Post("/request-code", user.RequestCode(gorm, mailer))
	apiUserGroup.Post("/verify-email", user.VerifyEmail(gorm))

	apiUserGroup.Get("/domains", user.Domains(gorm))

	apiAuthGroup := apiGroup.Group("/auth")
	apiAuthGroup.Post("/refresh", apiauth.RefreshToken(jwtHandler, gorm))

}
