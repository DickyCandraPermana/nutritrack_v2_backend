package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/MyFirstGo/internal/domain"
)

type Storage struct {
	Users interface {
		GetPaginated(context.Context, int, int) ([]*domain.User, error)
		GetAll(context.Context) ([]domain.User, error)
		GetByID(context.Context, int64) (*domain.User, error)
		GetByEmail(context.Context, string) (*domain.User, error)
		Create(context.Context, *domain.User) error
		Update(context.Context, *domain.User) error
		Delete(context.Context, int64) error
	}

	Foods interface {
		GetPaginated(context.Context, int, int) ([]*domain.Food, error)
		GetByID(context.Context, int64) (*domain.Food, error)
		Create(context.Context, *domain.Food) error
		Update(context.Context, *domain.Food) error
		Delete(context.Context, int64) error
	}

	Diary interface {
		GetSummary(context.Context, int64, time.Time) (*domain.DailySummary, error)
		GetEntries(context.Context, int64, time.Time) ([]*domain.FoodDiary, error)
		Create(context.Context, *domain.FoodDiary) error
		Delete(context.Context, int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
		Foods: &FoodStore{db},
		Diary: &DiaryStore{db},
	}
}
