package domain

import "time"

type User struct {
	ID            int64     `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	Height        float64   `json:"height"`
	Weight        float64   `json:"weight"`
	DateOfBirth   time.Time `json:"date_of_birth"`
	ActivityLevel int       `json:"activity_level"`
	Genter        string    `json:"gender"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

func (u *User) GetAge() int {
	today := time.Now()
	age := today.Year() - u.DateOfBirth.Year()

	if today.YearDay() < u.DateOfBirth.YearDay() {
		age--
	}
	return age
}
