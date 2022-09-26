package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/service/middleware"
	"github.com/maskrapp/backend/service/routes/api/email"
	"github.com/maskrapp/backend/service/routes/api/user"
	"github.com/supabase/postgrest-go"
)

func Setup(app *fiber.App, mailer *mailer.Mailer, postgrest *postgrest.Client, supabaseKey, supabaseBase string) {
	app.Use(cors.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})

	apiGroup := app.Group("/api")

	apiUserGroup := apiGroup.Group("/user")
	apiUserGroup.Use(middleware.AuthMiddleware(supabaseKey, supabaseBase))
	apiUserGroup.Post("/add-email", user.AddEmail(postgrest, mailer))
	apiUserGroup.Delete("/delete-email", user.DeleteEmail(postgrest))

	apiUserGroup.Post("/send-link", user.SendLink(postgrest, mailer))
	apiEmailGroup := apiGroup.Group("/email")
	apiEmailGroup.Post("/verify", email.VerifyEmail(postgrest))
}
