package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/middleware"
	"github.com/maskrapp/api/internal/routes/account"
	"github.com/maskrapp/api/internal/routes/auth"
	"github.com/maskrapp/api/internal/routes/auth/signin"
	"github.com/maskrapp/api/internal/routes/domains"
	"github.com/maskrapp/api/internal/routes/emails"
	"github.com/maskrapp/api/internal/routes/masks"
	"github.com/maskrapp/api/internal/routes/token"
)

func Setup(ctx global.Context, app *fiber.App) {
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})

	app.Get("/headers", func(c *fiber.Ctx) error {
		return c.JSON(c.GetReqHeaders())
	})

	signupGroup := app.Group("/auth/signup")

	signupGroup.Post("/", middleware.EmailRateLimit(ctx, 3, time.Minute, auth.Signup(ctx)))
	signupGroup.Post("/verify", middleware.EmailRateLimit(ctx, 3, time.Minute, auth.VerifySignup(ctx)))
	signupGroup.Post("/resend", middleware.EmailRateLimit(ctx, 3, time.Minute, auth.ResendCode(ctx)))
	signupGroup.Post("/create", middleware.EmailRateLimit(ctx, 5, time.Minute, auth.Create(ctx)))

	signinGroup := app.Group("/auth/signin")
	signinGroup.Post("/google", signin.Google(ctx))
	signinGroup.Post("/email", middleware.EmailRateLimit(ctx, 7, time.Minute, signin.Email(ctx)))

	resetPasswordGroup := app.Group("/auth/reset-password")
	resetPasswordGroup.Post("/", middleware.EmailRateLimit(ctx, 5, 5*time.Minute, auth.Reset(ctx)))
	resetPasswordGroup.Post("/verify", middleware.EmailRateLimit(ctx, 5, 5*time.Minute, auth.VerifyPassword(ctx)))
	resetPasswordGroup.Post("/confirm", auth.Confirm(ctx))

	emailsGroup := app.Group("/emails")
	emailsGroup.Use(middleware.AuthMiddleware(ctx))
	emailsGroup.Get("/", middleware.UserRateLimit(ctx, 30, time.Minute, emails.Get(ctx)))
	emailsGroup.Post("/new", middleware.UserRateLimit(ctx, 5, time.Minute, emails.Add(ctx)))
	emailsGroup.Delete("/:email", middleware.UserRateLimit(ctx, 15, time.Minute, emails.Delete(ctx)))
	emailsGroup.Post("/:email/verify", middleware.UserRateLimit(ctx, 15, time.Minute, emails.Verify(ctx)))
	emailsGroup.Post("/:email/create-code", middleware.UserRateLimit(ctx, 5, time.Minute, emails.RequestCode(ctx)))

	masksGroup := app.Group("/masks")
	masksGroup.Use(middleware.AuthMiddleware(ctx))
	masksGroup.Get("/", middleware.UserRateLimit(ctx, 30, time.Minute, masks.Get(ctx)))
	masksGroup.Post("/new", middleware.UserRateLimit(ctx, 5, time.Minute, masks.Add(ctx)))
	masksGroup.Delete("/:mask", middleware.UserRateLimit(ctx, 15, time.Minute, masks.Delete(ctx)))
	masksGroup.Put("/:mask/status", middleware.UserRateLimit(ctx, 15, time.Minute, masks.Status(ctx)))

	domainsGroup := app.Group("/domains")
	domainsGroup.Use(middleware.AuthMiddleware(ctx))
	domainsGroup.Get("/", middleware.UserRateLimit(ctx, 30, time.Minute, domains.Get(ctx)))

	accountGroup := app.Group("/account")
	accountGroup.Use(middleware.AuthMiddleware(ctx))
	accountGroup.Get("/", account.Get(ctx))

	tokenGroup := app.Group("/token")
	tokenGroup.Post("/refresh", token.Refresh(ctx))
	tokenGroup.Post("/revoke", token.Revoke(ctx))

}
