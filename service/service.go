package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/service/routes"
	"github.com/supabase/postgrest-go"
)

type BackendService struct {
	fiber     *fiber.App
	mailer    *mailer.Mailer
	postgrest *postgrest.Client
}

func New(emailToken, templateToken, postgrestURL, postgrestToken string, production bool) *BackendService {
	service := &BackendService{
		mailer:    mailer.New(emailToken, templateToken, production),
		fiber:     fiber.New(),
		postgrest: postgrest.NewClient(postgrestURL+"/rest/v1", "public", map[string]string{}).TokenAuth(postgrestToken),
	}
	if service.postgrest.ClientError != nil {
		panic(service.postgrest.ClientError)
	}
	routes.Setup(service.fiber, service.mailer, service.postgrest, postgrestToken, postgrestURL)
	return service
}

func (b *BackendService) Start() {
	b.fiber.Listen(":80")
}

func (b *BackendService) Shutdown() {
	b.fiber.Shutdown()
}
