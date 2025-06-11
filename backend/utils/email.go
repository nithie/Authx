package utils

import (
	"fmt"
	"log"
	"net/smtp"
)

func SendEmailVerificationToken(toEmail string, token string) (bool, error) {
	from := GetEnv("EMAIL", "")
	appPassword := GetEnv("EMAIL_PASS", "")

	log.Println(appPassword)

	smptpHost := "localhost"
	smtpPort := "1025"

	verifyLink := fmt.Sprintf("http://localhost:3000/verify?token=%s", token)

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
