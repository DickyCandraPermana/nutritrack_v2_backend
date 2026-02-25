package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MyFirstGo/internal/helper"
)

// Gunakan custom type untuk context key agar tidak bentrok dengan library lain
type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		userID, err := helper.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Simpan userID ke context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// Lanjut ke handler berikutnya dengan context baru
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
