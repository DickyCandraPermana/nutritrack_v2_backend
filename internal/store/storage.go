package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
	}
	Users interface {
		GetAll(context.Context) ([]User, error)
		GetByID(context.Context, int64) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		Create(context.Context, *User) error
		Update(context.Context, *User) error
		Delete(context.Context, int64) error
	}
	Foods interface {
		GetPaginated(context.Context, int, int) ([]Food, error)
		GetByID(context.Context, int64) (*Food, error)
		Create(context.Context, *Food) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostStore{db},
		Users: &UserStore{db},
		Foods: &FoodStore{db},
	}
}
