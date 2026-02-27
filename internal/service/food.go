package service

import (
	"context"
	"fmt"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/mapper"
	"github.com/MyFirstGo/internal/store"
	"github.com/MyFirstGo/pkg/converter"
	"github.com/go-playground/validator/v10"
)

type foodService struct {
	store     store.Storage
	validator validator.Validate
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

func (s *foodService) Create(ctx context.Context, input *domain.CreateFoodInput) (*domain.Food, error) {

	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	servingSize, servingUnit := float64(100), "g"

	if input.ServingSize != nil {
		input.ServingSize = &servingSize
	}

	if input.ServingUnit != nil {
		input.ServingUnit = &servingUnit
	}

	food := mapper.CreateFoodInputToFood(input)

	s.validateFoodNutrients(*food)

	for _, n := range input.Nutrients {
		food.Nutrients = append(food.Nutrients, domain.NutrientAmount{
			ID:     n.ID,
			Name:   n.Name,
			Unit:   n.Unit,
			Amount: n.Amount,
		})
	}

	if err := s.store.Foods.Create(ctx, food); err != nil {
		return nil, err
	}

	return food, nil
}

func (s *foodService) Update(ctx context.Context, id int64, input domain.UpdateFoodInput) (*domain.Food, error) {
	// 1. Ambil data asli dari DB
	food, err := s.store.Foods.GetByID(ctx, id)
	if err != nil {
		return nil, err // Pastikan store return ErrNotFound jika tidak ada
	}

	// 2. Patching: Update field hanya jika user mengirimkan datanya (tidak nil)
	if input.Name != nil {
		food.Name = *input.Name
	}
	if input.Description != nil {
		food.Description = *input.Description
	}
	if input.ServingSize != nil {
		food.ServingSize = input.ServingSize
	}
	if input.ServingUnit != nil {
		food.ServingUnit = input.ServingUnit
	}

	// 3. Logic Update Nutrients (Replace strategy)
	if input.Nutrients != nil {
		newNutrients := make([]domain.NutrientAmount, 0, len(*input.Nutrients))
		for _, n := range *input.Nutrients {
			newNutrients = append(newNutrients, domain.NutrientAmount{
				ID:     n.ID,
				Amount: n.Amount,
			})
		}
		food.Nutrients = newNutrients
	}

	// 4. Jalankan validasi bisnis (misal: kalori tidak boleh negatif)
	if err := s.validateFoodNutrients(*food); err != nil {
		return nil, err
	}

	// 5. Simpan ke Store
	if err := s.store.Foods.Update(ctx, food); err != nil {
		return nil, err
	}

	return food, nil
}

func (s *foodService) Delete(ctx context.Context, id int64) error {
	return s.store.Foods.Delete(ctx, id)
}
