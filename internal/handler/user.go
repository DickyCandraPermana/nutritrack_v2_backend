package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	App *app.Application
}

func NewUserHandler(app *app.Application) *UserHandler {
	return &UserHandler{App: app}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.App.Store.Users.GetAll(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []store.User{}
	}

	h.App.WriteJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.App.WriteJSON(w, http.StatusBadRequest, "Invalid User ID format")
		return
	}

	ctx := r.Context()

	user, err := h.App.Store.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.WriteJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.App.WriteJSON(w, http.StatusOK, user); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
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
	if err := h.App.Store.Users.Create(ctx, user); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.App.WriteJSON(w, http.StatusCreated, user); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	ctx := r.Context()

	user, err := h.App.Store.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.WriteJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var payload struct {
		Username *string `json:"username" validate:"min=3,max=30"`
		Email    *string `json:"email" validate:"email"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.App.Validator.Struct(payload); err != nil {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if payload.Username != nil {
		user.Username = *payload.Username
	}
	if payload.Email != nil {
		user.Email = *payload.Email
	}

	if err := h.App.Store.Users.Update(ctx, user); err != nil {
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	ctx := r.Context()

	if _, err := h.App.Store.Users.GetByID(ctx, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.WriteJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.App.Store.Users.Delete(ctx, id); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
