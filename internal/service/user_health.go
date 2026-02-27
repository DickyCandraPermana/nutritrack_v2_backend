package service

import (
	"context"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/helper"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
)

type UserHealthService struct {
	store     store.Storage
	validator validator.Validate
}

func (s *UserHealthService) GetUserHealthSummary(ctx context.Context, userID int64) (*domain.UserHealthSum, error) {
	user, err := s.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sum := helper.GetUserSummary(user)

	return sum, nil
}
