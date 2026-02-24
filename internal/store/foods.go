package store

import (
	"context"
	"database/sql"

	"github.com/MyFirstGo/internal/domain"
	"github.com/lib/pq"
)

type FoodStore struct {
	db *sql.DB
}

func (s *FoodStore) GetPaginated(ctx context.Context, limit, offset int) ([]*domain.Food, error) {
	queryFoods := `
	SELECT id, name, serving_size, serving_unit
	FROM foods
	WHERE deleted_at IS NULL
	LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, queryFoods, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foods []*domain.Food
	var foodIDs []int64
	foodMap := make(map[int64]*domain.Food)

	for rows.Next() {
		f := &domain.Food{Nutrients: []domain.NutrientAmount{}}
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.ServingSize, &f.ServingUnit); err != nil {
			return nil, err
		}

		foods = append(foods, f)
		foodIDs = append(foodIDs, f.ID)
		foodMap[f.ID] = f
	}

	if len(foodIDs) == 0 {
		return foods, nil
	}

	queryNutrient := `
	SELECT fn.food_id, n.id, n.name, n.unit, fn.amount
	FROM food_nutrients fn
	JOIN nutrients n ON fn.nutrient_id = n.id
	WHERE fn.food_id = ANY($1)
	`

	nutRows, err := s.db.QueryContext(ctx, queryNutrient, pq.Array(foodIDs))
	if err != nil {
		return nil, err
	}
	defer nutRows.Close()

	for nutRows.Next() {
		var foodID int64
		var na domain.NutrientAmount
		if err := nutRows.Scan(&foodID, &na.ID, &na.Name, &na.Unit, &na.Amount); err != nil {
			return nil, err
		}

		if f, ok := foodMap[foodID]; ok {
			f.Nutrients = append(f.Nutrients, na)
		}
	}

	return nil, nil
}

func (s *FoodStore) GetByID(ctx context.Context, id int64) (*domain.Food, error) {
	query := `
        SELECT f.id,
							 f.name,
							 f.description,
							 f.serving_size,
							 f.serving_unit,
							 fn.amount,
							 n.name AS nutrient_name,
							 n.unit,
							 f.created_at,
							 f.updated_at
        FROM foods f
				JOIN food_nutrients fn ON fn.food_id = f.id
				JOIN nutrients n ON n.id = fn.nutrient_id
        WHERE f.id = $1
					AND deleted_at IS NULL
    `

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var food *domain.Food

	for rows.Next() {
		var nID sql.NullInt64
		var nName, nUnit sql.NullString
		var nAmount sql.NullFloat64

		if food == nil {
			food = &domain.Food{Nutrients: []domain.NutrientAmount{}}
			err = rows.Scan(
				&food.ID, &food.Name, &food.Description, &food.ServingSize, &food.ServingUnit,
				&nID, &nName, &nUnit, &nAmount,
			)
		} else {
			var ignoreID int64
			var ignoreName, ignoreUnit, ignoreDescription string
			var ignoreSize float64
			err = rows.Scan(
				&ignoreID, &ignoreName, &ignoreDescription, &ignoreSize, &ignoreUnit,
				&nID, &nName, &nUnit, &nAmount,
			)
		}

		if err != nil {
			return nil, err
		}

		if nID.Valid {
			food.Nutrients = append(food.Nutrients, domain.NutrientAmount{
				ID:     nID.Int64,
				Name:   nName.String,
				Unit:   nUnit.String,
				Amount: nAmount.Float64,
			})
		}
	}

	if food == nil {
		return nil, ErrNotFound
	}

	return food, nil
}

func (s *FoodStore) Create(ctx context.Context, food *domain.Food) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	queryFood := `
			INSERT INTO foods (name, description, serving_size, serving_unit)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at
	`

	err = tx.QueryRowContext(ctx, queryFood,
		food.Name,
		food.Description,
		food.ServingSize,
		food.ServingUnit,
	).Scan(&food.ID, &food.CreatedAt)

	if err != nil {
		return err
	}

	if len(food.Nutrients) > 0 {
		queryNutrient := `
		INSERT INTO food_nutrients (food_id, nutrient_id, amount)
		VALUES ($1, $2, $3)
		`
		stmt, err := tx.PrepareContext(ctx, queryNutrient)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, n := range food.Nutrients {
			_, err := stmt.ExecContext(ctx, food.ID, n.ID, n.Amount)
			if err != nil {
				return nil
			}
		}
	}

	return tx.Commit()
}

func (s *FoodStore) Update(ctx context.Context, food *domain.Food) error {
	// 1. Mulai Transaksi
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. Update data utama makanan
	queryFood := `
		UPDATE foods
		SET name = $1, description = $2, serving_size = $3, serving_unit = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL`

	res, err := tx.ExecContext(ctx, queryFood,
		food.Name,
		food.Description,
		food.ServingSize,
		food.ServingUnit,
		food.ID,
	)
	if err != nil {
		return err
	}

	// Cek apakah ada baris yang diupdate (takutnya ID tidak ketemu)
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	if food.Nutrients != nil {
		// 3. Hapus nutrisi lama (Clean up)
		queryDeleteNutrients := `DELETE FROM food_nutrients WHERE food_id = $1`
		if _, err := tx.ExecContext(ctx, queryDeleteNutrients, food.ID); err != nil {
			return err
		}

		// 4. Insert nutrisi baru (Re-insert)
		if len(food.Nutrients) > 0 {
			queryInsert := `INSERT INTO food_nutrients (food_id, nutrient_id, amount) VALUES ($1, $2, $3)`
			stmt, err := tx.PrepareContext(ctx, queryInsert)
			if err != nil {
				return err
			}
			defer stmt.Close()

			for _, n := range food.Nutrients {
				if _, err := stmt.ExecContext(ctx, food.ID, n.ID, n.Amount); err != nil {
					return err
				}
			}
		}
	}

	// 5. Selesaikan Transaksi
	return tx.Commit()
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
