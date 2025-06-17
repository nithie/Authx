package utils

import (
	"errors"
	"net/http"
	"os"
	"strings"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ExtractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("Missing token")
	}

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
		return "", errors.New("invalid token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	return tokenString, nil

}
