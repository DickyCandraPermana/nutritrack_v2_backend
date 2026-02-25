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
