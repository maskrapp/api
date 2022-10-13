package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/maskrapp/backend/service"
)

func main() {
	service := service.New(os.Getenv("MAIL_TOKEN"), os.Getenv("TEMPLATE_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_DATABASE"), os.Getenv("PRODUCTION") == "true")
	service.Start()
}
