package main

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/maskrapp/backend/internal/service"
	"github.com/maskrapp/backend/internal/utils"
	"github.com/sirupsen/logrus"
)

func main() {

	routeEnvs := map[string]string{
		"/api/user/request-code":    "REQUEST_CODE_RPM",
		"/api/user/add-mask":        "ADD_MASK_RPM",
		"/api/user/add-email":       "ADD_EMAIL_RPM",
		"/api/user/delete-mask":     "DELETE_MASK_RPM",
		"/api/user/delete-email":    "DELETE_EMAIL_RPM",
		"/api/user/masks":           "MASKS_RPM",
		"/api/user/emails":          "EMAILS_RPM",
		"/api/user/set-mask-status": "SET_MASK_STATUS_RPM",
		"/api/user/verify-email":    "VERIFY_EMAIL_RPM",
	}

	routeRPMS := make(map[string]int)
	globalRPI := utils.ConvertWithFallback(os.Getenv("GLOBAL_RPI"), 50)

	for path, envValue := range routeEnvs {
		if envValue != "" {
			envValue := os.Getenv(envValue)
			intValue, err := strconv.Atoi(envValue)
			if err == nil {
				routeRPMS[path] = intValue
			}
		}
	}

	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		lvl = "debug"
	}
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	logrus.SetLevel(ll)

	service := service.New(os.Getenv("MAIL_TOKEN"), os.Getenv("TEMPLATE_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_DATABASE"), os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PASSWORD"), os.Getenv("CAPTCHA_SECRET"), os.Getenv("PRODUCTION") == "true", globalRPI, routeRPMS)
	service.Start()

}
