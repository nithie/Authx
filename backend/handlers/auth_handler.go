package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"

	"github.com/nithiee/authx/config"
	"github.com/nithiee/authx/models"
	"github.com/nithiee/authx/utils"
)

type SignupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	VerificationToken string `json:"verification_token"`
	Password          string `json:"password"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var input SignupInput
	json.NewDecoder(r.Body).Decode(&input)

	ok, err := utils.ValidatePasswordString(input.Password)

	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	go func() {
		_, err := utils.SendEmailVerificationToken(user.Email, verifyToken)
		if err != nil {
			log.Println("Sending verification email failed")
		}
	}()

	json.NewEncoder(w).Encode(map[string]string{"message": "Signup Successful"})
}

func Signin(w http.ResponseWriter, r *http.Request) {
	var input SignupInput
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

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)

	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
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
	var input ResetPasswordInput
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
	var input ForgotPasswordInput
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
