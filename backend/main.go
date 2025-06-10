package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nithiee/authx/config"
	"github.com/nithiee/authx/models"
	"github.com/nithiee/authx/routes"
)

func main() {
	config.Init()
	config.InitRedis()
	config.DB.AutoMigrate(&models.User{})

	r := chi.NewRouter()
	r.Mount("/auth", routes.AuthRoutes())
	r.Mount("/user", routes.UserRoutes())

	port := os.Getenv("PORT")
	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
