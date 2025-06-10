package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nithiee/authx/config"
	"github.com/nithiee/authx/middleware"
	"github.com/nithiee/authx/models"
	"github.com/nithiee/authx/utils"
)

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ResendVerificationLinkInput struct {
	Email string `json:"email"`
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	userIdVal := r.Context().Value(middleware.UserIDKey)

	if userIdVal == nil {
		http.Error(w, "Unauthoried", http.StatusUnauthorized)
		return
	}

	userID, ok := userIdVal.(uint)

	if !ok {
		http.Error(w, "Invalid user", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.Password = ""

	json.NewEncoder(w).Encode(user)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err1 := utils.ExtractToken(r)

	if err1 != nil {
		http.Error(w, "Logout Failed", http.StatusInternalServerError)
		return
	}

	err := config.RedisClient.Set(config.RedisContext, tokenString, "true", time.Hour*24).Err()

	if err != nil {
		http.Error(w, "Logout Failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var changePasswordInput ChangePasswordInput
	json.NewDecoder(r.Body).Decode(&changePasswordInput)
	userIdVal := r.Context().Value(middleware.UserIDKey)

	if userIdVal == nil {
		http.Error(w, "Unauthoried", http.StatusUnauthorized)
		return
	}

	userId, ok := userIdVal.(uint)

	if !ok {
		http.Error(w, "Invalid user id", http.StatusInternalServerError)
		return
	}

	if changePasswordInput.CurrentPassword == "" {
		http.Error(w, "Current Password cannot be empty", http.StatusInternalServerError)
		return
	}

	if changePasswordInput.NewPassword == "" {
		http.Error(w, "New Password cannot be empty", http.StatusInternalServerError)
		return
	}

	if changePasswordInput.NewPassword == "" {
		http.Error(w, "New Password cannot be empty", http.StatusInternalServerError)
		return
	}

	var user models.User

	ok, err := utils.ValidatePasswordString(changePasswordInput.NewPassword)

	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := config.DB.First(&user, userId).Error; err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if !utils.CheckPassword(user.Password, changePasswordInput.CurrentPassword) {
		http.Error(w, "Current password doesn't mactch", http.StatusInternalServerError)
		return
	}

	hashedPwd, err := utils.HashPassword(changePasswordInput.NewPassword)

	if err != nil {
		http.Error(w, "Error while hashing the password", http.StatusInternalServerError)
		return
	}

	if hashedPwd == user.Password {
		http.Error(w, "New password cannot be same as old password", http.StatusInternalServerError)
		return
	}

	user.Password = hashedPwd

	if err := config.DB.Save(user).Error; err != nil {
		http.Error(w, "Error while updating the password", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfuly"})

}
