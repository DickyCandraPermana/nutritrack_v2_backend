package service

import (
	"context"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	Auth interface {
		Login(context.Context, domain.UserLoginInput) (*domain.LoginResponse, error)
	}
	Users interface {
		GetPaginated(context.Context, int, int) ([]*domain.User, error)
		GetByID(context.Context, int64) (*domain.User, error)
		GetByEmail(context.Context, string) (*domain.User, error)
		Create(context.Context, domain.UserCreateInput) (*domain.UserResponse, error)
		Update(context.Context, int64, domain.UserUpdateInput) (*domain.UserResponse, error)
		Delete(context.Context, int64) error
	}

	Foods interface {
		GetPaginated(context.Context, int, int) ([]*domain.Food, error)
		GetByID(context.Context, int64) (*domain.Food, error)
		Create(context.Context, *domain.Food) error
		Update(context.Context, *domain.Food) error
		Delete(context.Context, int64) error
	}
}

func NewService(store store.Storage, validator validator.Validate) Service {
	return Service{
		Auth:  &AuthService{store, validator},
		Users: &UserService{store, validator},
		Foods: &foodService{store},
	}
}
