package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"redditclone/pkg/session/repository/redis"
	"redditclone/tools"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func ValidateJWTToken(repo *redis.SessionRedisManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenKey := []byte(os.Getenv("TOKEN_KEY"))
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			tools.JSONError(w, http.StatusUnauthorized, "missing token", "middleware.ValidateJWTToken")
			return
		}

		fieldParts := strings.Split(tokenString, " ")
		if len(fieldParts) != 2 || fieldParts[0] != "Bearer" {
			tools.JSONError(w, http.StatusUnauthorized, "bad token format", "middleware.ValidateJWTToken")
			return
		}
		pureToken := fieldParts[1]

		token, err := jwt.Parse(pureToken, func(token *jwt.Token) (interface{}, error) {
			method, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok || method.Alg() != "HS256" {
				return nil, errors.New("bad sign method")
			}
			return tokenKey, nil
		})
		if err != nil || !token.Valid {
			tools.JSONError(w, http.StatusUnauthorized, err.Error()+" | bad token", "middleware.ValidateJWTToken")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			tools.JSONError(w, http.StatusUnauthorized, "no payload", "middleware.ValidateJWTToken")
			return
		}

		claimsUser := claims["user"].(map[string]interface{})
		userIDString := claimsUser["id"].(string)
		userID, err := strconv.Atoi(userIDString)
		if err != nil {
			tools.JSONError(w, http.StatusInternalServerError, "type cast error", "middleware.ValidateJWTToken")
			return
		}
		_, err = repo.Check(userID)
		if err != nil {
			tools.JSONError(w, http.StatusUnauthorized, "no session", "SessionRedisManager.Check")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
