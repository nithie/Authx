package utils

import (
	"errors"
	"regexp"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidatePasswordString(password string) (bool, error) {
	if utf8.RuneCountInString(password) < 8 {
		return false, errors.New("Password should be greater than 8")
	}

	// Check for at least one uppercase character
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return false, errors.New("Password should contain alteast one uppercase character")
	}

	// Check for at least one lowercase character
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return false, errors.New("Password should contain alteast one lowercase character")
	}

	// Check for at least one numerical character
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return false, errors.New("Password should contain alteast one numeric character")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return false, errors.New("password must contain at least one special character")
	}

	return true, nil

}
