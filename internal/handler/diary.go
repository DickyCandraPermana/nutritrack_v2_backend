package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type DiaryHandler struct {
	App *app.Application
}

func NewDiaryHandler(app *app.Application) *DiaryHandler {
	return &DiaryHandler{
		App: app,
	}
}

func (h *DiaryHandler) GetDiariesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	dateStr := r.URL.Query().Get("date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		fmt.Printf("Error printing date %s", dateStr)
		date = time.Now()
	}

	ctx := r.Context()

	entries, err := h.App.Service.Diary.GetSummaryByUserId(ctx, userID, date)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, entries)
}

func (h *DiaryHandler) GetDiaryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	diaryIDStr := chi.URLParam(r, "diaryID")
	diaryID, err := strconv.ParseInt(diaryIDStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	diary, err := h.App.Service.Diary.GetDiaryWithUserId(r.Context(), userID, diaryID)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, diary)
}

func (h *DiaryHandler) CreateLogHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	// Gunakan string untuk tanggal di payload
	var payload struct {
		FoodID         int64   `json:"food_id"`
		AmountConsumed float64 `json:"amount_consumed"`
		ConsumedAt     string  `json:"consumed_at"` // Ubah jadi string
		MealType       string  `json:"meal_type"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	// Parse string ke time.Time manual
	var consumedAt time.Time
	var err error

	if payload.ConsumedAt != "" {
		// Coba parse format YYYY-MM-DD
		consumedAt, err = time.Parse("2006-01-02", payload.ConsumedAt)
		if err != nil {
			// Jika gagal, coba format RFC3339 (siapa tahu user kirim lengkap)
			consumedAt, err = time.Parse(time.RFC3339, payload.ConsumedAt)
			if err != nil {
				h.App.BadRequestResponse(w, r, errors.New("invalid date format, use YYYY-MM-DD"))
				return
			}
		}
	} else {
		consumedAt = time.Now() // Default ke sekarang jika kosong
	}

	input := &domain.DiaryCreateInput{
		UserID:         userID,
		FoodID:         payload.FoodID,
		AmountConsumed: payload.AmountConsumed,
		ConsumedAt:     consumedAt, // Sekarang sudah bertipe time.Time
		MealType:       payload.MealType,
	}

	diary, err := h.App.Service.Diary.Create(r.Context(), input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			h.App.ValidationErrorResponse(w, r, err)
			return
		}

		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, diary)
}

func (h *DiaryHandler) UpdateLogHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	diaryIDStr := chi.URLParam(r, "diaryID")
	diaryID, err := strconv.ParseInt(diaryIDStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	var payload struct {
		AmountConsumed *float64 `json:"amount_consumed"`
		ConsumedAt     *string  `json:"consumed_at"`
		MealType       *string  `json:"meal_type"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	// 1. Buat variabel pointer untuk menampung hasil parsing
	var finalTime *time.Time

	// 2. Cek apakah field 'consumed_at' ada di JSON
	if payload.ConsumedAt != nil {
		var t time.Time
		var err error

		if *payload.ConsumedAt == "" {
			// Jika user kirim string kosong "", kita set ke jam sekarang
			t = time.Now()
		} else {
			// Coba parse YYYY-MM-DD
			t, err = time.Parse("2006-01-02", *payload.ConsumedAt)
			if err != nil {
				// Jika gagal, coba RFC3339
				t, err = time.Parse(time.RFC3339, *payload.ConsumedAt)
				if err != nil {
					h.App.BadRequestResponse(w, r, errors.New("invalid date format"))
					return
				}
			}
		}
		// Ambil alamat dari hasil parsing
		finalTime = &t
	}

	// 3. Masukkan ke input.
	// Jika payload.ConsumedAt tadi nil, maka finalTime juga nil.
	input := &domain.DiaryUpdateInput{
		ID:             diaryID,
		AmountConsumed: payload.AmountConsumed,
		ConsumedAt:     finalTime, // Ini akan nil jika tidak dikirim di JSON
		MealType:       payload.MealType,
	}

	diary, err := h.App.Service.Diary.Update(r.Context(), userID, input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			h.App.ValidationErrorResponse(w, r, err)
			return
		}

		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, diary)
}

func (h *DiaryHandler) DeleteLogHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	diaryIDStr := chi.URLParam(r, "diaryID")
	diaryID, err := strconv.ParseInt(diaryIDStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	if err = h.App.Service.Diary.Delete(r.Context(), userID, diaryID); err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusNoContent, nil)

}
