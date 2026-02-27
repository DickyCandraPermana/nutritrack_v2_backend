package domain

type Food struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ServingSize *float64         `json:"serving_size"`
	ServingUnit *string          `json:"serving_unit"`
	Nutrients   []NutrientAmount `json:"nutrients"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

type NutrientAmount struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Unit   string  `json:"unit"`
	Amount float64 `json:"amount"`
}

type FoodFilter struct {
	Query       string
	MinCalories float64
	MaxCalories float64
	Limit       int
	Offset      int
}

type CreateFoodInput struct {
	Name        string   `validate:"required"`
	Description string   `validate:"omitempty"`
	ServingSize *float64 `validate:"omitempty"`
	ServingUnit *string  `validate:"omitempty"`
	Nutrients   []struct {
		ID     int64   `validate:"required"`
		Name   string  `validate:"required"`
		Unit   string  `validate:"required"`
		Amount float64 `validate:"required"`
	} `validate:"omitempty"`
}

type UpdateFoodInput struct {
	Name        *string
	Description *string
	ServingSize *float64
	ServingUnit *string
	Nutrients   *[]UpdateNutrientInput
}

type UpdateNutrientInput struct {
	ID     int64
	Amount float64
}
