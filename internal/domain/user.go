package domain

import (
	"time"
)

type User struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	Height        *float64   `json:"height"`
	Weight        *float64   `json:"weight"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	ActivityLevel *int       `json:"activity_level"`
	Gender        *string    `json:"gender"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

func (u *User) GetAge() int {
	today := time.Now()
	age := today.Year() - u.DateOfBirth.Year()

	if today.YearDay() < u.DateOfBirth.YearDay() {
		age--
	}
	return age
}

// func (u *User) GetBMI() string {
// 	heightInMeters := u.Height / 100
// 	bmi := u.Weight / math.Pow(heightInMeters, 2)

// 	var result string

// 	switch {
// 	case bmi < 17.5:
// 		result = "kurus"
// 	case bmi >= 17.5 && bmi < 23:
// 		result = "normal"
// 	case bmi >= 23 && bmi < 25:
// 		result = "gemuk"
// 	case bmi >= 25 && bmi < 30:
// 		result = "obesitas i"
// 	case bmi >= 30:
// 		result = "obesitas ii"
// 	}

// 	return result
// }

// func (u *User) GetBMR() float64 {
// 	var result float64

// 	var genVar float64

// 	switch u.Gender {
// 	case "male":
// 		genVar = 5
// 	case "female":
// 		genVar = -161
// 	}

// 	result = (10 * u.Weight) + (6.25 * u.Height) - (5 * float64(u.GetAge())) + genVar

// 	return result
// }
