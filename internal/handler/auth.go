package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/store"
)

type AuthHandler struct {
	App *app.Application
}

func NewAuthHandler(app *app.Application) *AuthHandler {
	return &AuthHandler{App: app}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	input := domain.UserLoginInput{
		Email:    payload.Email,
		Password: payload.Password,
	}

	res, err := h.App.Service.Auth.Login(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			h.App.WriteJSON(w, http.StatusUnauthorized, "email atau password salah", nil)
		case errors.Is(err, domain.ErrInvalidCredentials):
			h.App.WriteJSON(w, http.StatusUnauthorized, "email atau password salah", nil)
		default:
			log.Printf("Database error in LoginHandler: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	// Success response
	h.App.WriteJSON(w, http.StatusOK, res, nil)
}
