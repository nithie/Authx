package utils

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/nithiee/authx/internal/config"
)

func SendEmailVerificationToken(toEmail string, token string) (bool, error) {
	from := config.SmtpEmail
	appPassword := GetEnv("EMAIL_PASS", "")

	log.Println(appPassword)

	smptpHost := config.SmtpHost
	smtpPort := config.SmtpPort

	verifyLink := fmt.Sprintf("%s/verify?token=%s", config.AppUrl, token)

	subject := "Email verification"

	body := fmt.Sprintf("Click the link to verify your email: %s", verifyLink)

	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)

	auth := smtp.PlainAuth("", "", "", smptpHost)

	err := smtp.SendMail(smptpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)

	if err != nil {
		// fmt.Println("Failed to send verification mail:", err)
		return false, err
	}

	return true, nil
}
