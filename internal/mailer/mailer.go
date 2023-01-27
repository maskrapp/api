package mailer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/maskrapp/backend/internal/config"
)

type Mailer struct {
	httpClient   *req.Client
	token        string
	templateKey  string
	emailAddress string
	production   bool
}

func New(config *config.Config) *Mailer {
	httpClient := req.C()
	return &Mailer{
		httpClient:   httpClient,
		token:        config.ZeptoMail.EmailToken,
		templateKey:  config.ZeptoMail.TemplateKey,
		emailAddress: config.ZeptoMail.EmailAddress,
		production:   config.Production,
	}
}

func (m *Mailer) createJSON(email, code string) ([]byte, error) {
	// https://www.zoho.com/zeptomail/help/api/email-templates.html

	reqMap := make(map[string]interface{})
	reqMap["mail_template_key"] = m.templateKey
	reqMap["bounce_address"] = "bounce@bounce.maskr.org"
	reqMap["from"] = map[string]string{
		"address": m.emailAddress,
		"from":    "no-reply",
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

// SendVerifyEmail is used when a user adds a new email to their account.
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

// SendUserVerificationMail is used when a user creates their account.
func (m *Mailer) SendUserVerificationMail(email, code string) error {

	//TODO: use different template

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

func (m *Mailer) SendPasswordCodeMail(email, code string) error {

	//TODO: use different template

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
