package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nithiee/authx/internal/handlers"
)

func AuthRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/signup", handlers.Signup)
	r.Post("/signin", handlers.Signin)
	r.Get("/verify", handlers.VerifyHandler)
	r.Post("/resend-verify-link", handlers.ResendVerificationLink)
	r.Post("/forgot-password", handlers.ForgotPassword)
	r.Post("/reset-password", handlers.ResetPassword)
	return r
}
