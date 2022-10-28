package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Mailer struct {
	token       string
	templateKey string
	production  bool
}

func New(token string, templateKey string, production bool) *Mailer {
	return &Mailer{
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
	client := http.DefaultClient
	request, err := http.NewRequest("POST", "https://api.zeptomail.eu/v1.1/email/template", bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	authHeader := fmt.Sprintf("Zoho-enczapikey %v", m.token)
	request.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {authHeader},
	}
	resp, err := client.Do(request)
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if resp.StatusCode != 201 {
		errorMessage := fmt.Sprintf("expected status code 201, got: %v with response body: %v", resp.StatusCode, res)
		return errors.New(errorMessage)
	}
	return err
}

func (m *Mailer) SendUserVerificationMail(email, code string) error {

	data, err := m.createJSON(email, code)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	request, err := http.NewRequest("POST", "https://api.zeptomail.eu/v1.1/email/template", bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	authHeader := fmt.Sprintf("Zoho-enczapikey %v", m.token)
	request.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {authHeader},
	}
	resp, err := client.Do(request)
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if resp.StatusCode != 201 {
		errorMessage := fmt.Sprintf("expected status code 201, got: %v with response body: %v", resp.StatusCode, res)
		return errors.New(errorMessage)
	}
	return err
}
