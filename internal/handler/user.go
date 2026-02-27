package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/middleware"
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

func (h *UserHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		size = 10
	}
	ctx := r.Context()

	users, err := h.App.Service.Users.GetPaginated(ctx, page, size)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, users, nil)
}

func (h *UserHandler) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := h.App.Service.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.BadRequestResponse(w, r, err)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.App.WriteJSON(w, http.StatusOK, user, nil); err != nil {
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

	if err := h.App.WriteJSON(w, http.StatusCreated, user, nil); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) UpdateAvatarHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		h.App.BadRequestResponse(w, r, errors.New("file too large"))
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}
	defer file.Close()

	path, err := h.App.Service.Users.UpdateAvatar(r.Context(), userID, file)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]string{"url": path}, nil)
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
		Username      *string    `json:"username"`
		Email         *string    `json:"email"`
		Height        *float64   `json:"height"`
		Weight        *float64   `json:"weight"`
		DateOfBirth   *time.Time `json:"date_of_birth"`
		ActivityLevel *int       `json:"activity_level"`
		Gender        *string    `json:"gender"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	input := domain.UserUpdateInput{
		Email:         payload.Email,
		Username:      payload.Username,
		Height:        payload.Height,
		Weight:        payload.Weight,
		DateOfBirth:   payload.DateOfBirth,
		ActivityLevel: payload.ActivityLevel,
		Gender:        payload.Gender,
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
			h.App.BadRequestResponse(w, r, err)
			return
		}
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, user, nil)
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
			h.App.NotFoundResponse(w, r)
		case errors.Is(err, domain.ErrCannotDelete):
			h.App.WriteJSON(w, http.StatusConflict, map[string]string{
				"error": "User cannot be deleted (has associated data)",
			}, nil)
		default:
			log.Printf("Error deleting user %d: %v", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
