package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nithiee/authx/internal/config"
)

var jwtSecret = []byte(config.JwtSecret)

func GenerateAccessToken(userID uint, email string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	})

	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken() (string, error) {
	token, err := GenetateVerificationToken()
	return token, err
}
