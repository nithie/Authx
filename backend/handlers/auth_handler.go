package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nithiee/authx/config"
	"github.com/nithiee/authx/models"
	"github.com/nithiee/authx/utils"
)

type SignupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

func ResetPassword(w http.ResponseWriter, r *http.Request) {

}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {

}
