package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

func (m *Mailer) createJSON(email, name, code string) ([]byte, error) {
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

	link := "localhost:3000/verify/" + code
	if m.production {
		link = "alpha.maskr.app/verify/" + code
	}
	reqMap["merge_info"] = map[string]string{
		"name": name,
		"link": link,
	}
	return json.Marshal(reqMap)
}

func (m *Mailer) SendVerifyMail(email, name, code string) error {

	data, err := m.createJSON(email, name, code)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	request, err := http.NewRequest("POST", "https://api.zeptomail.eu/v1.1/email/template", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {m.token},
	}
	resp, err := client.Do(request)
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if resp.StatusCode != 201 {
		return errors.New("Expected status code 201, got " + strconv.Itoa(resp.StatusCode))
	}
	return err
}
