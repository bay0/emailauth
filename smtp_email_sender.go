package emailauth

import (
	"net/smtp"
)

type SMTPEmailSender struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	senderEmail  string
}

func NewSMTPEmailSender(smtpHost, smtpPort, smtpUsername, smtpPassword, senderEmail string) *SMTPEmailSender {
	return &SMTPEmailSender{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		senderEmail:  senderEmail,
	}
}

func (s *SMTPEmailSender) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.senderEmail, []string{to}, msg)
	if err != nil {
		return err
	}

	return nil
}
