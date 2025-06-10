package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email             string `gorm:"unique;not null" json:"email"`
	Password          string `gorm:"not null" json:"password,omitempty"`
	Role              string `gorm:"default:user"`
	IsVerified        bool   `json:"is_verified"`
	VerificationToken string `json:"verification_token"`
}
