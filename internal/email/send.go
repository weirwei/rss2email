package email

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type Email struct {
	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
}

type Config struct {
	Host string
	Port int
	User string
	Pass string
}

func NewEmail(config *Config) *Email {
	return &Email{
		smtpHost: config.Host,
		smtpPort: config.Port,
		smtpUser: config.User,
		smtpPass: config.Pass,
	}
}

func (e *Email) Send(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.smtpUser)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.smtpHost, e.smtpPort, e.smtpUser, e.smtpPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d.DialAndSend(m)
}
