package recaptcha

import (
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

type Recaptcha struct {
	httpClient *req.Client
	secret     string
}

func New(secret string) *Recaptcha {
	return &Recaptcha{req.C().DevMode(), secret}
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

	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	responseBody := &responseBody{}
	resp, err := r.httpClient.R().
		SetRetryFixedInterval(500 * time.Millisecond).
		SetRetryCondition(func(resp *req.Response, err error) bool {
			return !resp.IsSuccess()
		}).
		SetHeaders(headers).
    SetResult(responseBody).
		Post(url)



	if err != nil {
		logrus.Errorf("http request failed: %v", err)
		return false
	}

	if !resp.IsSuccess() {
		logrus.Errorf("http request failed: %v", resp.Dump())
		return false
	}

	if len(responseBody.ErrorCodes) > 0 {
		logrus.Error("captcha error codes:", responseBody.ErrorCodes)
	}

	if !responseBody.Success || responseBody.Action != action {
		return false
	}

	//according to https://stackoverflow.com/a/52170635, anything below 0.5 is malicious
	if responseBody.Score < 0.5 {
		return false
	}
	return true
}
