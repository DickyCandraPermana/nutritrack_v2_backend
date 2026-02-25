package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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
		users = []domain.User{}
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

	user, err := h.App.Service.Users.GetByID(ctx, id)
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
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	input := domain.UserCreateInput{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	}

	ctx := r.Context()
	user, err := h.App.Service.Users.Create(ctx, input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			h.App.ValidationErrorResponse(w, r, err)
			return
		}
		if errors.Is(err, domain.ErrDuplicateEmail) {
			http.Error(w, "Email sudah digunakan", http.StatusConflict)
			return
		}

		log.Printf("Unexpected error creating user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.App.WriteJSON(w, http.StatusCreated, user); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Error parsing url", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var payload struct {
		Username *string `json:"username"`
		Email    *string `json:"email"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	input := domain.UserUpdateInput{
		Email:    payload.Email,
		Username: payload.Username,
	}
	user, err := h.App.Service.Users.Update(ctx, id, input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			h.App.ValidationErrorResponse(w, r, err)
			return
		}
		if errors.Is(err, domain.ErrDuplicateEmail) {
			http.Error(w, "Email sudah digunakan", http.StatusConflict)
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			h.App.WriteJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := h.App.Service.Users.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			h.App.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		case errors.Is(err, domain.ErrCannotDelete):
			h.App.WriteJSON(w, http.StatusConflict, map[string]string{
				"error": "User cannot be deleted (has associated data)",
			})
		default:
			log.Printf("Error deleting user %d: %v", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
