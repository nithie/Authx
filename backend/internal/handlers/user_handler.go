package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nithiee/authx/internal/config"
	"github.com/nithiee/authx/internal/middleware"
	"github.com/nithiee/authx/internal/models"
	"github.com/nithiee/authx/pkg/utils"
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
	cookie, err := r.Cookie("refresh_token")

	if err != nil {
		http.Error(w, "Logout Failed", http.StatusInternalServerError)
		return
	}

	config.RedisClient.Del(config.RedisContext, fmt.Sprintf("refresh_token:%s", cookie.Value))

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true, // keep consistent with when you set it
		SameSite: http.SameSiteStrictMode,
	})

	tokenString, err := utils.ExtractToken(r)
	if err != nil {
		http.Error(w, "Logout Failed", http.StatusInternalServerError)
		return
	}

	err = config.RedisClient.Set(config.RedisContext, tokenString, "true", time.Hour*24).Err()

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

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")

	if err != nil {
		http.Error(w, "Missing Token", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value

	userIdStr, err := config.RedisClient.Get(config.RedisContext, fmt.Sprintf("refresh_tokem:%s", refreshToken)).Result()

	if err != nil {
		http.Error(w, "Invalid or Expired Refresh token", http.StatusUnauthorized)
		return
	}

	userID, _ := strconv.Atoi(userIdStr)

	var user models.User

	if err := config.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	accesToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	newRefreshToken, err := utils.GenerateRefreshToken()

	if err != nil {
		http.Error(w, "Failed to generate new refresh token", http.StatusInternalServerError)
		return
	}

	config.RedisClient.Del(config.RedisContext, fmt.Sprintf("refresh_tokem:%s", refreshToken))

	err = config.RedisClient.Set(config.RedisContext, fmt.Sprintf("refresh_token:%s", newRefreshToken), userID, 24*7*time.Hour).Err()

	if err != nil {
		http.Error(w, "Failed to store new refresh token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Secure:   true,
		HttpOnly: true,
		Path:     "/api/refresh",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accesToken,
	})
}
