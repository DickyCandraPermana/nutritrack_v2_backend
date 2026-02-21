package main

import (
	"fmt"
	"net/http"

	"github.com/MyFirstGo/internal/auth"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email" validate:"email,required"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := app.validator.Struct(payload); err != nil {
		var errDetails []string
		for _, err := range err.(validator.ValidationErrors) {
			errDetails = append(errDetails, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}

		app.writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":   "Validation failed",
			"details": errDetails,
		})
		return
	}

	ctx := r.Context()

	user, err := app.store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		app.writeJSON(w, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	fmt.Println(user.Password)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {

		app.writeJSON(w, http.StatusUnauthorized, fmt.Sprintf("%s is %s", user.Password, []byte(payload.Password)))
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Gagal generate token", http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
