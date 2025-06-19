package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nithiee/authx/internal/config"
	"github.com/nithiee/authx/pkg/utils"
)

type contextKey string

var secret = []byte(utils.GetEnv("JWT_SECRET", "supersecretkey"))

const UserIDKey contextKey = "userID"
const UserRoleKey contextKey = "userRole"

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString, err := utils.ExtractToken(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		isTokenBlackListed, _ := config.RedisClient.Get(config.RedisContext, tokenString).Result()

		if isTokenBlackListed == "true" {
			http.Error(w, "Token has been blacklisted", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(w, "Invalid climas", http.StatusUnauthorized)
			return
		}

		userId, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid userId", http.StatusUnauthorized)
			return
		}
		userRole, ok := claims["role"].(string)

		if !ok {
			http.Error(w, "Invalid userRole", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()

		ctx = context.WithValue(ctx, UserIDKey, uint(userId))
		ctx = context.WithValue(ctx, UserRoleKey, string(userRole))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
