package handler

import (
	"fmt"
	"net/http"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/auth"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	App *app.Application
}

func NewAuthHandler(app *app.Application) *AuthHandler {
	return &AuthHandler{App: app}
}

// TODO: Selesaikan copy dari cmd/api
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email" validate:"email,required"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.App.Validator.Struct(payload); err != nil {
		var errDetails []string
		for _, err := range err.(validator.ValidationErrors) {
			errDetails = append(errDetails, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}

		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":   "Validation failed",
			"details": errDetails,
		})
		return
	}

	ctx := r.Context()

	user, err := h.App.Store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		h.App.WriteJSON(w, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	fmt.Println(user.Password)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {

		h.App.WriteJSON(w, http.StatusUnauthorized, fmt.Sprintf("%s is %s", user.Password, []byte(payload.Password)))
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Gagal generate token", http.StatusInternalServerError)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
