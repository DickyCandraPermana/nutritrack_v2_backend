package service

import (
	"context"
	"fmt"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/store"
	"github.com/MyFirstGo/pkg/converter"
)

type foodService struct {
	store store.Storage
}

func (s *foodService) validateFoodNutrients(food domain.Food) error {
	servingSize, servingUnit := 100.00, "g"

	if food.ServingSize != nil {
		servingSize = *food.ServingSize
	}

	if food.ServingUnit != nil {
		servingUnit = *food.ServingUnit
	}

	totalWeightInGrams := converter.ToGrams(servingSize, servingUnit)

	for _, n := range food.Nutrients {
		if n.Amount < 0 {
			return fmt.Errorf("nilai nutrisi %s (%.2f %s) tidak boleh kurang dari 0!",
				n.Name, n.Amount, n.Unit)
		}

		if n.Unit == "kcal" {
			continue
		}

		nutrientInGrams := converter.ToGrams(n.Amount, n.Unit)

		if nutrientInGrams > totalWeightInGrams {
			return fmt.Errorf("nilai nutrisi %s (%.2fg) tidak boleh lebih dari total berat saji (%.2fg)!",
				n.Name, nutrientInGrams, totalWeightInGrams)
		}
	}

	return nil
}

func (s *foodService) GetPaginated(ctx context.Context, page, size int) ([]*domain.Food, error) {

	if page < 1 {
		page = 1
	}

	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	return s.store.Foods.GetPaginated(ctx, size, offset)
}

func (s *foodService) GetByID(ctx context.Context, id int64) (*domain.Food, error) {
	return s.store.Foods.GetByID(ctx, id)
}

func (s *foodService) Create(ctx context.Context, food *domain.Food) error {

	servingSize, servingUnit := float64(100), "g"

	if food.ServingSize != nil {
		food.ServingSize = &servingSize
	}

	if food.ServingUnit != nil {
		food.ServingUnit = &servingUnit
	}

	s.validateFoodNutrients(*food)

	return s.store.Foods.Create(ctx, food)
}

func (s *foodService) Update(ctx context.Context, food *domain.Food) error {
	s.validateFoodNutrients(*food)

	return s.store.Foods.Update(ctx, food)
}

func (s *foodService) Delete(ctx context.Context, id int64) error {
	return s.store.Foods.Delete(ctx, id)
}
