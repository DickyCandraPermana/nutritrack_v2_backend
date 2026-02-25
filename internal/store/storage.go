package store

import (
	"context"
	"database/sql"

	"github.com/MyFirstGo/internal/domain"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
	}
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
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostStore{db},
		Users: &UserStore{db},
		Foods: &FoodStore{db},
	}
}
