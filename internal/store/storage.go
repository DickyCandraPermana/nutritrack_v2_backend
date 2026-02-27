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
		UpdateAvatar(context.Context, int64, string) error
		Delete(context.Context, int64) error
	}

	Foods interface {
		Search(context.Context, domain.FoodFilter) ([]*domain.Food, error)
		GetPaginated(context.Context, int, int) ([]*domain.Food, error)
		GetByID(context.Context, int64) (*domain.Food, error)
		Create(context.Context, *domain.Food) error
		Update(context.Context, *domain.Food) error
		Delete(context.Context, int64) error
	}

	Diary interface {
		GetSummary(context.Context, int64, time.Time) (*domain.DailySummary, error)
		GetEntries(context.Context, int64, time.Time) ([]*domain.FoodDiary, error)
		GetUserEntry(context.Context, int64, int64) (*domain.FoodDiary, error)
		GetEntry(context.Context, int64) (*domain.FoodDiary, error)
		Create(context.Context, *domain.FoodDiary) error
		Update(context.Context, *domain.FoodDiary) error
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
