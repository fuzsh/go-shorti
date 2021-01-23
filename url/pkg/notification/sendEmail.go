package notification

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// EmailDetails have the data we need to send email.
type EmailDetails struct {
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	ContentType string   `json:"content_type"`
	Body        string   `json:"body"`
}

// SendEmail sends the email.
func SendEmail(details EmailDetails) error {
	url := "http://127.0.0.1:8084/api/v1"
	json, err := json.Marshal(details)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
