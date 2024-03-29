package email

import (
	"fmt"
	"net/smtp"
	"os"
)

type mailInterface interface {
	sendMail(to, subject, message string) error
}

type mail struct{}

func (m *mail) sendMail(to, subject, message string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("APP_PASSWORD")
	smtphost := os.Getenv("SMTP_HOST")
	smtpport := os.Getenv("SMTP_PORT")

	smtporigin := fmt.Sprintf("%s:%s", smtphost, smtpport)

	auth := smtp.PlainAuth(
		"",
		from,
		password,
		smtphost,
	)

	body := fmt.Sprintf("TO: %s\r\nSubject: %s\r\n\n\n%s", to, subject, message)

	return smtp.SendMail(
		smtporigin,
		auth,
		from,
		[]string{to},
		[]byte(body),
	)
}
