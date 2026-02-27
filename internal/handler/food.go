package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/helper"
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
	q := r.URL.Query()

	filter := domain.FoodFilter{
		Query:       q.Get("q"),
		MinCalories: helper.ReadFloatQuery(r, "min_cal", 0),
		MaxCalories: helper.ReadFloatQuery(r, "max_cal", 0),
		Limit:       helper.ReadIntQuery(r, "limit", 10),
		Offset:      (helper.ReadIntQuery(r, "page", 1) - 1) * 10,
	}

	foods, err := h.App.Service.Foods.Search(r.Context(), filter)
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
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ServingSize float64 `json:"serving_size"`
		ServingUnit string  `json:"serving_unit"`
		Nutrients   []struct {
			ID     int64   `json:"id"`
			Name   string  `json:"name"`
			Unit   string  `json:"unit"`
			Amount float64 `json:"amount"`
		} `json:"nutrients"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	input := &domain.CreateFoodInput{
		Name:        payload.Name,
		Description: payload.Description,
		ServingSize: &payload.ServingSize,
		ServingUnit: &payload.ServingUnit,
	}

	for _, n := range payload.Nutrients {
		input.Nutrients = append(input.Nutrients, struct {
			ID     int64   `validate:"required"`
			Name   string  `validate:"required"`
			Unit   string  `validate:"required"`
			Amount float64 `validate:"required"`
		}{
			ID:     n.ID,
			Name:   n.Name,
			Unit:   n.Unit,
			Amount: n.Amount,
		})
	}

	food, err := h.App.Service.Foods.Create(r.Context(), input)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, food)
}

func (h *FoodHandler) UpdateFoodsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "foodID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil { // Reintroduce this check
		h.App.BadRequestResponse(w, r, err)
		return
	}
	// Payload lokal untuk mapping JSON
	var payload struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		ServingSize *float64 `json:"serving_size"`
		ServingUnit *string  `json:"serving_unit"`
		Nutrients   *[]struct {
			ID     int64   `json:"id"`
			Amount float64 `json:"amount"`
		} `json:"nutrients"`
	}

	if err := h.App.ReadJSON(w, r, &payload); err != nil {
		h.App.BadRequestResponse(w, r, err)
		return
	}

	// Map payload ke Domain Input
	input := domain.UpdateFoodInput{
		Name:        payload.Name,
		Description: payload.Description,
		ServingSize: payload.ServingSize,
		ServingUnit: payload.ServingUnit,
	}

	if payload.Nutrients != nil {
		nutrients := make([]domain.UpdateNutrientInput, 0, len(*payload.Nutrients))
		for _, n := range *payload.Nutrients {
			nutrients = append(nutrients, domain.UpdateNutrientInput{
				ID:     n.ID,
				Amount: n.Amount,
			})
		}
		input.Nutrients = &nutrients
	}

	// Panggil Service. Service yang bertanggung jawab ambil data lama & update.
	food, err := h.App.Service.Foods.Update(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			h.App.NotFoundResponse(w, r)
		default:
			h.App.ServerErrorResponse(w, r, err)
		}
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
