package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getFoodsHandler(w http.ResponseWriter, r *http.Request) {
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

	offset := (page - 1) * size

	foods, err := app.store.Foods.GetPaginated(r.Context(), size, offset)
	if err != nil {
		app.serverErrorResponse(w, r, err) // Log error & kirim 500
		return
	}

	app.writeJSON(w, http.StatusOK, foods)
}

func (app *application) getFoodByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid food ID format"))
		return
	}

	food, err := app.store.Foods.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, food)
}

func (app *application) createFoodsHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string          `json:"name" validate:"required,max=255"`
		Description string          `json:"description"`
		Nutrients   store.Nutrients `json:"nutrients"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.validator.Struct(payload); err != nil {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, err.Error()) // Simpelnya kirim err.Error()
		return
	}

	food := &store.Food{
		Name:        payload.Name,
		Description: payload.Description,
		Nutrients:   payload.Nutrients,
	}

	if err := app.store.Foods.Create(r.Context(), food); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, food)
}

func (app *application) updateFoodsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var payload struct {
		Name        *string          `json:"name" validate:"omitempty,max=255"`
		Description *string          `json:"description"`
		Nutrients   *store.Nutrients `json:"nutrients"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	food, err := app.store.Foods.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	if payload.Name != nil {
		food.Name = *payload.Name
	}
	if payload.Description != nil {
		food.Description = *payload.Description
	}
	if payload.Nutrients != nil {
		food.Nutrients = *payload.Nutrients
	}

	if err := app.store.Foods.Update(ctx, food); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, food)
}

func (app *application) deleteFoodsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Foods.Delete(r.Context(), id); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
