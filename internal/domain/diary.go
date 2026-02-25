package domain

import "time"

type FoodDiary struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	FoodID         int64      `json:"food_id"`
	AmountConsumed float64    `json:"amount_consumed"`
	ConsumedAt     time.Time  `json:"consumed_at"`
	MealType       string     `json:"meal_type"`
	FoodName       string     `json:"food_name,omitempty"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// Summary untuk Dashboard
type DailySummary struct {
	TotalCalories float64     `json:"total_calories"`
	TotalProtein  float64     `json:"total_protein"`
	TotalCarbs    float64     `json:"total_carbs"`
	TotalFat      float64     `json:"total_fat"`
	Entries       []FoodDiary `json:"entries"`
}
