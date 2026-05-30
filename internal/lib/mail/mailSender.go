package mail

import (
	"Etog/internal/config"
	"fmt"
	"net/smtp"
)

func SendMail(data config.MailData, recipient, message, theme string) error {
	auth := smtp.PlainAuth(
		"",
		data.From,
		data.Password,
		data.Host,
	)
	to := []string{recipient}
	msg := []byte(
		"From: " + data.From + "\r\n" +
			"To: " + recipient + "\r\n" +
			"Subject: " + theme + "\r\n" +
			"\r\n" +
			message,
	)
	return smtp.SendMail(fmt.Sprintf("%s%s%s", data.Host, ":", data.Port), auth, data.From, to, msg)
}
