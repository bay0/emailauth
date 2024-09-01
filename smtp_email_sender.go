package emailauth

import (
	"fmt"
	"net/smtp"
)

type SMTPEmailSender struct {
	smtpHost         string
	smtpPort         string
	smtpUsername     string
	smtpPassword     string
	senderEmail      string
	allowUnencrypted bool
}

func NewSMTPEmailSender(smtpHost, smtpPort, smtpUsername, smtpPassword, senderEmail string, allowUnencrypted bool) *SMTPEmailSender {
	return &SMTPEmailSender{
		smtpHost:         smtpHost,
		smtpPort:         smtpPort,
		smtpUsername:     smtpUsername,
		smtpPassword:     smtpPassword,
		senderEmail:      senderEmail,
		allowUnencrypted: allowUnencrypted,
	}
}

func (s *SMTPEmailSender) SendEmail(to, subject, body string) error {
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))

	smtpAddr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	var auth smtp.Auth
	if !s.allowUnencrypted {
		auth = smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	}

	var err error
	if s.allowUnencrypted {
		// For unencrypted connections (like MailHog in development)
		err = smtp.SendMail(smtpAddr, nil, s.senderEmail, []string{to}, msg)
	} else {
		// For encrypted connections (production use)
		err = smtp.SendMail(smtpAddr, auth, s.senderEmail, []string{to}, msg)
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
