package service

import (
	"context"
	"time"

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
		UpdatePassword(context.Context, int64, string) (*domain.UserResponse, error)
		Delete(context.Context, int64) error
	}

	Diary interface {
		GetSummaryByUserId(context.Context, int64, time.Time) (*domain.DailySummary, error)
		GetDiaryByDiaryId(context.Context, int64) (*domain.FoodDiary, error)
		GetDiaryWithUserId(context.Context, int64, int64) (*domain.FoodDiary, error)
		Create(context.Context, *domain.DiaryCreateInput) (*domain.FoodDiary, error)
		Update(context.Context, int64, *domain.DiaryUpdateInput) (*domain.FoodDiary, error)
		Delete(context.Context, int64, int64) error
	}

	Foods interface {
		GetPaginated(context.Context, int, int) ([]*domain.Food, error)
		GetByID(context.Context, int64) (*domain.Food, error)
		Create(context.Context, *domain.Food) error
		Update(context.Context, *domain.Food) error
		Delete(context.Context, int64) error
	}

	Health interface {
		GetUserHealthSummary(context.Context, int64) (*domain.UserHealthSum, error)
	}
}

func NewService(store store.Storage, validator validator.Validate) Service {
	return Service{
		Auth:   &AuthService{store, validator},
		Users:  &UserService{store, validator},
		Diary:  &DiaryService{store, validator},
		Foods:  &foodService{store},
		Health: &UserHealthService{store, validator},
	}
}
