package auth

import (
	"time"

	"github.com/MyFirstGo/internal/env"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(env.GetString("JWT_SECRET", "secret"))

func GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
