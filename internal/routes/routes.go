package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/middleware"
	apiauth "github.com/maskrapp/backend/internal/routes/api/auth"
	"github.com/maskrapp/backend/internal/routes/api/user"
	"github.com/maskrapp/backend/internal/routes/auth"
)

func Setup(ctx global.Context, app *fiber.App) {
	app.Use(cors.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})

	app.Get("/headers", func(c *fiber.Ctx) error {
		return c.JSON(c.GetReqHeaders())
	})

	authGroup := app.Group("/auth")
	authGroup.Post("/google", auth.GoogleHandler(ctx))
	authGroup.Post("create-account-code", middleware.EmailRateLimit(ctx, 3, time.Minute, auth.CreateAccountCode(ctx)))
	authGroup.Post("verify-account-code", middleware.EmailRateLimit(ctx, 5, time.Minute, auth.VerifyAccountCode(ctx)))
	authGroup.Post("resend-account-code", middleware.EmailRateLimit(ctx, 3, time.Minute, auth.ResendAccountCode(ctx)))
	authGroup.Post("create-account", middleware.EmailRateLimit(ctx, 5, time.Minute, auth.CreateAccount(ctx)))
	authGroup.Post("email-login", middleware.EmailRateLimit(ctx, 7, time.Minute, auth.EmailLogin(ctx)))

	authGroup.Post("create-password-code", middleware.EmailRateLimit(ctx, 5, 5*time.Minute, auth.CreatePasswordCode(ctx)))
	authGroup.Post("verify-password-code", middleware.EmailRateLimit(ctx, 5, 5*time.Minute, auth.VerifyPasswordCode(ctx)))
	authGroup.Post("change-password", auth.ChangePassword(ctx))

	apiGroup := app.Group("/api")

	apiUserGroup := apiGroup.Group("/user")
	apiUserGroup.Use(middleware.AuthMiddleware(ctx))

	apiUserGroup.Get("/account-details", middleware.UserRateLimit(ctx, 30, time.Minute, user.AccountDetails(ctx)))

	apiUserGroup.Post("/emails", middleware.UserRateLimit(ctx, 30, time.Minute, user.Emails(ctx)))
	apiUserGroup.Post("/add-email", middleware.UserRateLimit(ctx, 5, time.Minute, user.AddEmail(ctx)))
	apiUserGroup.Delete("/delete-email", middleware.UserRateLimit(ctx, 15, time.Minute, user.DeleteEmail(ctx)))

	apiUserGroup.Post("/masks", middleware.UserRateLimit(ctx, 30, time.Minute, user.Masks(ctx)))
	apiUserGroup.Post("add-mask", middleware.UserRateLimit(ctx, 5, time.Minute, user.AddMask(ctx)))
	apiUserGroup.Delete("delete-mask", middleware.UserRateLimit(ctx, 15, time.Minute, user.DeleteMask(ctx)))
	apiUserGroup.Put("set-mask-status", middleware.UserRateLimit(ctx, 15, time.Minute, user.SetMaskStatus(ctx)))

	apiUserGroup.Post("/request-code", middleware.UserRateLimit(ctx, 5, time.Minute, user.RequestCode(ctx)))
	apiUserGroup.Post("/verify-email", middleware.UserRateLimit(ctx, 15, time.Minute, user.VerifyEmail(ctx)))
	apiUserGroup.Get("/domains", middleware.UserRateLimit(ctx, 30, time.Minute, user.Domains(ctx)))

	apiAuthGroup := apiGroup.Group("/auth")
	apiAuthGroup.Post("/refresh", apiauth.RefreshToken(ctx))
	apiAuthGroup.Post("/revoke-token", apiauth.RevokeToken(ctx))
}
