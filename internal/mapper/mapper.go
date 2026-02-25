package mapper

import "github.com/MyFirstGo/internal/domain"

func UserToUserResponse(user *domain.User) *domain.UserResponse {
	res := &domain.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	if user.Weight != nil {
		res.Weight = user.Weight
	}

	if user.Height != nil {
		res.Height = user.Height
	}

	if user.DateOfBirth != nil {
		res.DateOfBirth = user.DateOfBirth
	}

	if user.ActivityLevel != nil {
		res.ActivityLevel = user.ActivityLevel
	}

	if user.Gender != nil {
		res.Gender = user.Gender
	}

	return res
}
