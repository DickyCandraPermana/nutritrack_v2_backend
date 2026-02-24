package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
)

type FoodHandler struct {
	App *app.Application
}

func NewFoodHandler(app *app.Application) *FoodHandler {
	return &FoodHandler{App: app}
}

func (h *FoodHandler) GetFoodsHandler(w http.ResponseWriter, r *http.Request) {
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

	foods, err := h.App.Store.Foods.GetPaginated(r.Context(), size, offset)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err) // Log error & kirim 500
		return
	}

	h.App.WriteJSON(w, http.StatusOK, foods)
}

func (h *FoodHandler) GetFoodByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, fmt.Errorf("invalid food ID format"))
		return
	}

	food, err := h.App.Store.Foods.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.NotFoundResponse(w, r)
			return
		}
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, food)
}

func (h *FoodHandler) CreateFoodsHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string          `json:"name" validate:"required,max=255"`
		Description string          `json:"description"`
		Nutrients   store.Nutrients `json:"nutrients"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	if err := h.App.Validator.Struct(payload); err != nil {
		h.App.ErrorResponse(w, r, http.StatusUnprocessableEntity, err.Error()) // Simpelnya kirim err.Error()
		return
	}

	food := &store.Food{
		Name:        payload.Name,
		Description: payload.Description,
		Nutrients:   payload.Nutrients,
	}

	if err := h.App.Store.Foods.Create(r.Context(), food); err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, food)
}

func (h *FoodHandler) UpdateFoodsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	var payload struct {
		Name        *string          `json:"name" validate:"omitempty,max=255"`
		Description *string          `json:"description"`
		Nutrients   *store.Nutrients `json:"nutrients"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	food, err := h.App.Store.Foods.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.App.NotFoundResponse(w, r)
			return
		}
		h.App.ServerErrorResponse(w, r, err)
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

	if err := h.App.Store.Foods.Update(ctx, food); err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, food)
}

func (h *FoodHandler) DeleteFoodsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	if err := h.App.Store.Foods.Delete(r.Context(), id); err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
