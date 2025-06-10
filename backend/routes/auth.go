package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nithiee/authx/handlers"
)

func AuthRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/signup", handlers.Signup)
	r.Post("/signin", handlers.Signin)

	return r
}
