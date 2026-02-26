package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/middleware"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
)

type ProfileHandler struct {
	App *app.Application
}

func NewProfileHandler(app *app.Application) *ProfileHandler {
	return &ProfileHandler{
		App: app,
	}
}

func (h *ProfileHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	user, _ := h.App.Service.Users.GetByID(r.Context(), userID)

	// 3. Ambil Summary Nutrisi Hari Ini
	summary, _ := h.App.Service.Diary.GetSummaryByUserId(r.Context(), userID, time.Now())

	// 4. Gabungkan dalam satu response cantik
	h.App.WriteJSON(w, http.StatusOK, map[string]any{
		"user":    user,
		"summary": summary,
	})
}

func (h *ProfileHandler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

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

	ctx := r.Context()

	user, err := h.App.Service.Users.Update(ctx, userID, input)
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

func (h *ProfileHandler) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var payload struct {
		Password string `json:"password"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	user, err := h.App.Service.Users.UpdatePassword(r.Context(), userID, payload.Password)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.WriteJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, user)
}
