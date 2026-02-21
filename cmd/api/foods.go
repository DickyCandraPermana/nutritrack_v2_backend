package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func (app *application) getFoodsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil query param, kasih default value kalau kosong
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	size, _ := strconv.Atoi(sizeStr)
	if size < 1 || size > 100 {
		size = 10
	}

	// 2. Hitung offset
	offset := (page - 1) * size

	foods, err := app.store.Foods.GetPaginated(r.Context(), size, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, foods)
}

func (app *application) createFoodsHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string          `json:"name" validate:"required,max=255"`
		Description string          `json:"description"`
		Nutrients   store.Nutrients `json:"nutrients"`
	}

	// 1. Baca JSON
	if err := readJSON(w, r, &payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 2. Validasi (Jika kamu pakai library validator)
	if err := app.validator.Struct(payload); err != nil {
		var errDetails []string
		for _, err := range err.(validator.ValidationErrors) {
			errDetails = append(errDetails, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}

		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":   "Validation failed",
			"details": errDetails,
		})
		return
	}

	// 3. Siapkan struct Store
	food := &store.Food{
		Name:        payload.Name,
		Description: payload.Description,
		Nutrients:   payload.Nutrients, // Tipe datanya sudah cocok!
	}

	// 4. Simpan ke database
	ctx := r.Context()
	if err := app.store.Foods.Create(ctx, food); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, food)
}

func (app *application) getFoodByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, "Invalid User ID format: "+err.Error())
		return
	}

	ctx := r.Context()

	user, err := app.store.Foods.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, "User not found")
			return
		}
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusOK, user); err != nil {
		http.Error(w, "Internal Server Error"+err.Error(), http.StatusInternalServerError)
		return
	}
}
