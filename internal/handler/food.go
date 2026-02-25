package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
)

type FoodHandler struct {
	App *app.Application
}

func NewFoodHandler(app *app.Application) *FoodHandler {
	return &FoodHandler{
		App: app,
	}
}

func (h *FoodHandler) GetFoodsHandler(w http.ResponseWriter, r *http.Request) {
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

	foods, err := h.App.Service.Foods.GetPaginated(r.Context(), page, size)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err)
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

	food, err := h.App.Service.Foods.GetByID(r.Context(), id)
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
		Name        string  `json:"name" validate:"required,max=255"`
		Description string  `json:"description"`
		ServingSize float64 `json:"serving_size"`
		ServingUnit string  `json:"serving_unit"`
		Nutrients   []struct {
			ID     int64   `json:"id" validate:"required"`
			Name   string  `json:"name"`
			Unit   string  `json:"unit"`
			Amount float64 `json:"amount" validate:"required"`
		} `json:"nutrients"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	if err := h.App.Validator.Struct(payload); err != nil {
		h.App.ErrorResponse(w, r, http.StatusUnprocessableEntity, err.Error()) // Simpelnya kirim err.Error()
		return
	}

	food := &domain.Food{
		Name:        payload.Name,
		Description: payload.Description,
		ServingSize: &payload.ServingSize,
		ServingUnit: &payload.ServingUnit,
	}

	for _, n := range payload.Nutrients {
		food.Nutrients = append(food.Nutrients, domain.NutrientAmount{
			ID:     n.ID,
			Name:   n.Name,
			Unit:   n.Unit,
			Amount: n.Amount,
		})
	}

	if err := h.App.Service.Foods.Create(r.Context(), food); err != nil {
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
		Name        *string  `json:"name" validate:"max=255"`
		Description *string  `json:"description"`
		ServingSize *float64 `json:"serving_size"`
		ServingUnit *string  `json:"serving_unit"`
		Nutrients   *[]struct {
			ID     int64   `json:"id" validate:"required"`
			Amount float64 `json:"amount" validate:"required"`
		} `json:"nutrients"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	food, err := h.App.Service.Foods.GetByID(ctx, id)
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
	if payload.ServingSize != nil {
		food.ServingSize = payload.ServingSize
	}
	if payload.ServingUnit != nil {
		food.ServingUnit = payload.ServingUnit
	}
	if payload.Nutrients != nil {
		food.Nutrients = make([]domain.NutrientAmount, 0, len(*payload.Nutrients))

		for _, n := range *payload.Nutrients {
			food.Nutrients = append(food.Nutrients, domain.NutrientAmount{
				ID:     n.ID,
				Amount: n.Amount,
			})
		}
	}

	if err := h.App.Service.Foods.Update(ctx, food); err != nil {
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

	if err := h.App.Service.Foods.Delete(r.Context(), id); err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
