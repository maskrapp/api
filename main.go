package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/maskrapp/backend/service"
)

func main() {
	service := service.New(os.Getenv("MAIL_TOKEN"), os.Getenv("TEMPLATE_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("PRODUCTION") == "true", os.Getenv("POSTGRES_URI"))
	service.Start()
}
