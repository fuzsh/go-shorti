package lib

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

// Email struct hold data that need to send email
type Email struct {
	To          []string
	Subject     string
	Body        string
	ContentType string `json:"content_type"`
}

// Send sends the message to all the defined users
func (m *Email) Send() error {
	gm := gomail.NewMessage()
	gm.SetHeader("From", Conf.SMTP.User)
	gm.SetHeader("To", m.To...)
	gm.SetHeader("Subject", m.Subject)
	gm.SetBody(m.ContentType, m.Body)

	d := gomail.NewDialer(Conf.SMTP.Host, Conf.SMTP.Port, Conf.SMTP.User, Conf.SMTP.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	d.SSL = true
	err := d.DialAndSend(gm)

	return err
}
