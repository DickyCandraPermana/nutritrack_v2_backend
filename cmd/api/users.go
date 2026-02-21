package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := app.store.Users.GetAll(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []store.User{}
	}

	app.writeJSON(w, http.StatusOK, users)
}

func (app *application) getUserById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, "Invalid User ID format")
		return
	}

	ctx := r.Context()

	user, err := app.store.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.writeJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, user); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Gagal memproses password", http.StatusInternalServerError)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	ctx := r.Context()
	if err := app.store.Users.Create(ctx, user); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, user); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	ctx := r.Context()

	user, err := app.store.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.writeJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var payload struct {
		Username *string `json:"username" validate:"min=3,max=30"`
		Email    *string `json:"email" validate:"email"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := app.validator.Struct(payload); err != nil {
		app.writeJSON(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if payload.Username != nil {
		user.Username = *payload.Username
	}
	if payload.Email != nil {
		user.Email = *payload.Email
	}

	if err := app.store.Users.Update(ctx, user); err != nil {
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusOK, user)
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	ctx := r.Context()

	if _, err := app.store.Users.GetByID(ctx, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.writeJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := app.store.Users.Delete(ctx, id); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
