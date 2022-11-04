package recaptcha

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Recaptcha struct {
	httpClient *http.Client
	logger     *logrus.Logger
	secret     string
}

func New(httpClient *http.Client, logger *logrus.Logger, secret string) *Recaptcha {
	return &Recaptcha{httpClient, logger, secret}
}

type responseBody struct {
	Success            bool          `json:"success"`
	Score              float32       `json:"score"`
	Action             string        `json:"action"`
	ChallengeTimestamp string        `json:"challenge_ts"`
	Hostname           string        `json:"hostname"`
	ErrorCodes         []interface{} `json:"error-codes"`
}

func (r *Recaptcha) ValidateCaptchaToken(token, action string) bool {

	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%v&response=%v", r.secret, token)

	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		r.logger.Error("http request initialization failed: ", err)
		return false
	}
	request.Header = map[string][]string{
		"Accept":       {"application/json"},
		"Content-Type": {"application/json"},
	}

	response, err := r.httpClient.Do(request)
	if err != nil {
		r.logger.Error("http request failed: ", err)
		return false
	}
	defer response.Body.Close()

	resBody := &responseBody{}
	err = json.NewDecoder(response.Body).Decode(resBody)
	if err != nil {
		r.logger.Error("JSON decoder error: ", err)
		return false
	}

	if len(resBody.ErrorCodes) > 0 {
		r.logger.Error("captcha error codes:", resBody.ErrorCodes)
	}

	if !resBody.Success || resBody.Action != action {
		return false
	}

	//according to https://stackoverflow.com/a/52170635, anything below 0.5 is malicious
	if resBody.Score < 0.5 {
		return false
	}
	return true
}
