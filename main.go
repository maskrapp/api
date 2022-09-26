package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/maskrapp/backend/service"
)

func main() {
	service := service.New(os.Getenv("MAIL_TOKEN"), os.Getenv("TEMPLATE_KEY"), os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_TOKEN"), os.Getenv("PRODUCTION") == "true")
	service.Start()
}
