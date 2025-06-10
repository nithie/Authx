package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nithiee/authx/handlers"
	"github.com/nithiee/authx/middleware"
)

func UserRoutes() http.Handler {
	r := chi.NewRouter()

	r.Group(func(api chi.Router) {
		api.Use(middleware.JWTAuth)
		api.Use(middleware.RoleAuth("user"))
		api.Get("/me", handlers.MeHandler)
		api.Get("/logout", handlers.LogoutHandler)
		api.Post("/change-password", handlers.ChangePassword)
	})

	return r
}
