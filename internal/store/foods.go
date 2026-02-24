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
	WHERE deleted_at IS NULL
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
        WHERE id = $1 AND deleted_at IS NULL
    `

	food := &Food{}

	// 1. Deklarasikan variabel sementara untuk kolom nullable
	var description sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&food.ID,
		&food.Name,
		&description, // Scan ke sql.NullString
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

	// 2. Assign ke struct. Jika NULL, .String otomatis jadi "" (empty string)
	food.Description = description.String

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

func (s *FoodStore) Update(ctx context.Context, food *Food) error {
	query := `
	UPDATE foods
	SET
		name = $2,
		description = $3,
		nutrients = $4,
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	`

	res, err := s.db.ExecContext(ctx, query, food.ID, food.Name, food.Description, food.Nutrients)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if row == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *FoodStore) Delete(ctx context.Context, id int64) error {
	query := `
	UPDATE foods
	SET deleted_at = NOW()
	WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return ErrNotFound
	}

	return nil
}
