package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Food struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Nutrients   Nutrients `json:"nutrients"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

type Nutrients map[string]float64

// Value mengonversi map ke JSON saat simpan ke DB
func (n Nutrients) Value() (driver.Value, error) {
	return json.Marshal(n)
}

// Scan mengonversi JSON dari DB kembali ke map saat select
func (n *Nutrients) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &n)
}

type FoodStore struct {
	db *sql.DB
}

func (s *FoodStore) GetPaginated(ctx context.Context, limit, offset int) ([]Food, error) {
	query := `
	SELECT
		id,
		name,
		COALESCE(description, ''),
		nutrients,
		created_at,
		updated_at
	FROM foods
	ORDER BY id
	LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var foods []Food

	for rows.Next() {
		var food Food

		err := rows.Scan(
			&food.ID,
			&food.Name,
			&food.Description,
			&food.Nutrients,
			&food.CreatedAt,
			&food.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		foods = append(foods, food)
	}

	return foods, nil
}

func (s *FoodStore) GetByID(ctx context.Context, id int64) (*Food, error) {
	query := `
	SELECT id, name, description, nutrients, created_at, updated_at
	FROM foods
	WHERE id = $1
	`

	food := &Food{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&food.ID,
		&food.Name,
		&food.Description,
		&food.Nutrients,
		&food.CreatedAt,
		&food.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return food, nil
}

func (s *FoodStore) Create(ctx context.Context, food *Food) error {
	query := `
	INSERT INTO foods (name, description, nutrients)
	VALUES ($1, $2, $3) RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		food.Name,
		food.Description,
		food.Nutrients,
	).Scan(
		&food.ID,
		&food.CreatedAt,
		&food.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}
