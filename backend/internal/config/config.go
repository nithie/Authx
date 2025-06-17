package config

import "os"

var (
	DBUrl     string
	RedisUrl  string
	JwtSecret string
	Port      string
	SmtpEmail string
	SmtpHost  string
	SmtpPort  string
	AppUrl    string
)

func LoadEnv() {
	DBUrl = getEnv("DB_URL")
	RedisUrl = getEnv("REDIS_URL")
	JwtSecret = getEnv("JWT_SECRET")
	Port = getEnv("PORT")

	SmtpEmail = getEnv("EMAIL")
	SmtpHost = getEnv("SMTP_HOST")
	SmtpPort = getEnv("SMTP_PORT")
	AppUrl = getEnv("APP_URL")

}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return ""
}
