package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/nithiee/authx/internal/config"
	"github.com/nithiee/authx/internal/models"
	"github.com/nithiee/authx/internal/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env file not found or couldn't be loaded (proceeding with system env)")
	}

	config.LoadEnv()
	config.Init()
	config.InitRedis()
	config.DB.AutoMigrate(&models.User{})

	r := chi.NewRouter()
	r.Mount("/auth", routes.AuthRoutes())
	r.Mount("/user", routes.UserRoutes())

	port := config.Port
	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
