package mailer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

type Mailer struct {
	httpClient  *req.Client
	token       string
	templateKey string
	production  bool
}

func New(token string, templateKey string, production bool) *Mailer {
	httpClient := req.C()
	return &Mailer{
		httpClient:  httpClient,
		token:       token,
		templateKey: templateKey,
		production:  production,
	}
}

func (m *Mailer) createJSON(email, code string) ([]byte, error) {
	// https://www.zoho.com/zeptomail/help/api/email-templates.html

	reqMap := make(map[string]interface{})
	reqMap["mail_template_key"] = m.templateKey
	reqMap["bounce_address"] = "bounce@bounce.maskr.app"
	reqMap["from"] = map[string]string{
		"address": "no-reply@maskr.app",
		"from":    "maskr.app",
	}
	var recipients []interface{}
	recipient := map[string]interface{}{
		"email_address": map[string]string{
			"address": email,
			"name":    email,
		},
	}

	recipients = append(recipients, recipient)

	reqMap["to"] = recipients

	reqMap["merge_info"] = map[string]string{
		"code": code,
	}
	return json.Marshal(reqMap)
}

func (m *Mailer) SendVerifyMail(email, code string) error {

	data, err := m.createJSON(email, code)
	if err != nil {
		return err
	}
	var responseData map[string]interface{}
	headers := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Zoho-enczapikey %v", m.token),
	}
	resp, err := m.httpClient.R().SetBody(data).
		SetRetryCount(3).
		SetRetryFixedInterval(500 * time.Millisecond).
		SetRetryCondition(func(resp *req.Response, err error) bool {
			return !resp.IsSuccess()
		}).
		SetHeaders(headers).
		SetResult(responseData).
		Post("https://api.zeptomail.eu/v1.1/email/template")

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("expected status code 201, got: %v with response body: %v", resp.StatusCode, responseData)
	}
	return nil
}

func (m *Mailer) SendUserVerificationMail(email, code string) error {

	data, err := m.createJSON(email, code)
	if err != nil {
		return err
	}
	var responseData map[string]interface{}
	headers := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Zoho-enczapikey %v", m.token),
	}

	resp, err := m.httpClient.R().SetBody(data).
		SetRetryCount(3).
		SetRetryFixedInterval(500 * time.Millisecond).
		SetRetryCondition(func(resp *req.Response, err error) bool {
			return !resp.IsSuccess()
		}).
		SetHeaders(headers).
		SetResult(responseData).
		Post("https://api.zeptomail.eu/v1.1/email/template")

	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("expected status code 201, got: %v with response body: %v", resp.StatusCode, responseData)
	}
	return err
}
