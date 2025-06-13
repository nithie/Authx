package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte(GetEnv("JWT_SECRET", "supersecretkey"))

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
	token := uuid.NewString()
	return token, nil
}
