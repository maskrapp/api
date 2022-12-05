package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Database struct {
		Hostname string
		Username string
		Password string
		Database string
	}
	Redis struct {
		Host     string
		Username string
		Password string
	}
	Recaptcha struct {
		Secret string
	}
	ZeptoMail struct {
		EmailToken  string
		TemplateKey string
	}
	JWT struct {
		Secret string
	}
	OAuth struct {
		GoogleRedirectURL string
		GoogleClientId    string
		GoogleSecret      string
	}
	Logger struct {
		LogLevel string
	}
	Production bool
}

func New() *Config {
	cfg := &Config{}

	cfg.Database.Database = os.Getenv("POSTGRES_DATABASE")
	cfg.Database.Hostname = os.Getenv("POSTGRES_HOST")
	cfg.Database.Username = os.Getenv("POSTGRES_USER")
	cfg.Database.Password = os.Getenv("POSTGRES_PASSWORD")

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	cfg.Redis.Username = os.Getenv("REDIS_USERNAME")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	cfg.Recaptcha.Secret = os.Getenv("CAPTCHA_SECRET")

	cfg.ZeptoMail.EmailToken = os.Getenv("MAIL_TOKEN")
	cfg.ZeptoMail.TemplateKey = os.Getenv("MAIL_TEMPLATE")

	cfg.JWT.Secret = os.Getenv("SECRET_KEY")

	cfg.OAuth.GoogleClientId = os.Getenv("GOOGLE_CLIENT_ID")
	cfg.OAuth.GoogleRedirectURL = os.Getenv("GOOGLE_REDIRECT")
	cfg.OAuth.GoogleSecret = os.Getenv("GOOGLE_SECRET")

	cfg.Logger.LogLevel = getOrDefault("LOG_LEVEL", "debug")

	cfg.Production = getOrDefault("PRODUCTION", "true") == "true"

	return cfg
}

func getOrDefault(variable string, def string) string {
	result, ok := os.LookupEnv(variable)
	if !ok {
		return def
	}
	return result
}
