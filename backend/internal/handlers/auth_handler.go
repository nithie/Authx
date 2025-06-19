package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nithiee/authx/internal/config"
	"github.com/nithiee/authx/internal/models"
	"github.com/nithiee/authx/pkg/utils"
)

type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	VerificationToken string `json:"verification_token" validate:"required"`
	Password          string `json:"password" validate:"required,min=8"`
}

//	Signup godoc
//	@Summary Signup
//	@Tags Auth
//	@Accept json
//	@Produce json
//	@Router /api/v1/auth/signup [post]

func Signup(w http.ResponseWriter, r *http.Request) {
	var input SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid Json")
		return
	}

	validate := validator.New()

	if err := validate.Struct(input); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err := utils.ValidatePasswordString(input.Password)

	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hashedPwd, err := utils.HashPassword(input.Password)
	if err != nil {
		http.Error(w, " Failed to hash password", http.StatusInternalServerError)
		return
	}

	verifyToken, err := utils.GenetateVerificationToken()

	if err != nil {
		http.Error(w, "Failed to generate verification link", http.StatusInternalServerError)
		return
	}

	user := models.User{Email: input.Email, Password: hashedPwd, IsVerified: false, VerificationToken: verifyToken}

	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error creating user", http.StatusBadRequest)
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)

	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()

	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/refresh",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	err = config.RedisClient.Set(config.RedisContext, fmt.Sprintf("refresh_token_%s", refreshToken), user.ID, time.Hour*24).Err()

	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	go func() {
		_, err := utils.SendEmailVerificationToken(user.Email, verifyToken)
		if err != nil {
			log.Println("Sending verification email failed")
		}
	}()

	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}

func Signin(w http.ResponseWriter, r *http.Request) {
	var input SignupRequest
	if err := json.NewDecoder(r.Body).Decode((&input)); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User

	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPassword(user.Password, input.Password) {
		http.Error(w, "Inavalid password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)

	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()

	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/refresh",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	err = config.RedisClient.Set(config.RedisContext, fmt.Sprintf("refresh_token_%s", refreshToken), user.ID, time.Hour*24*7).Err()

	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": token,
	})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	token := queryParams.Get("token")

	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := config.DB.Where("verification_token = ?", token).First(&user).Error; err != nil {
		http.Error(w, "Invalid or Expired Token", http.StatusBadRequest)
		return
	}

	user.IsVerified = true
	user.VerificationToken = ""

	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully",
	})
}

func ResendVerificationLink(w http.ResponseWriter, r *http.Request) {
	var input ResendVerificationLinkInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := mail.ParseAddress(input.Email)

	if err != nil {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	token, err := utils.GenetateVerificationToken()

	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	go func() {
		_, err = utils.SendEmailVerificationToken(input.Email, token)
		if err != nil {
			log.Println("Sending verification email failed", err)
		}
	}()

	json.NewEncoder(w).Encode(map[string]string{"message": "Verification link sent successfuly"})
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := utils.ValidatePasswordString(input.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User

	if err := config.DB.Where("reset_password_token = ?", input.VerificationToken).First(&user).Error; err != nil {
		http.Error(w, "Inavalid Token", http.StatusBadRequest)
		return
	}

	hashedPwd, err := utils.HashPassword(input.Password)

	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.Password = hashedPwd
	user.ResetPasswordToken = ""

	if err := config.DB.Save(user).Error; err != nil {
		http.Error(w, "Failed to update the new password", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Successful"})

}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Inavalid request body", http.StatusBadRequest)
		return
	}

	_, err := mail.ParseAddress(input.Email)

	if err != nil {
		http.Error(w, "Inavalid email address", http.StatusBadRequest)
		return
	}

	token, err := utils.GenetateVerificationToken()

	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	user.ResetPasswordToken = token

	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	go func() {
		_, err = utils.SendEmailVerificationToken(input.Email, token)
		if err != nil {
			log.Println("Error while sending verification link")
		}
	}()

	json.NewEncoder(w).Encode(map[string]string{"message": "Verification link sent successfuly"})

}
