package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MyFirstGo/internal/domain"
)

type DiaryStore struct {
	db *sql.DB
}

func (s *DiaryStore) GetSummary(ctx context.Context, userID int64, date time.Time) (*domain.DailySummary, error) {
	query := `
        SELECT
            COALESCE(SUM(CASE WHEN n.name = 'Caloric Value' THEN (fd.amount_consumed / NULLIF(f.serving_size, 0)) * fn.amount END), 0) as calories,
            COALESCE(SUM(CASE WHEN n.name = 'Protein' THEN (fd.amount_consumed / NULLIF(f.serving_size, 0)) * fn.amount END), 0) as protein,
            COALESCE(SUM(CASE WHEN n.name = 'Carbohydrates' THEN (fd.amount_consumed / NULLIF(f.serving_size, 0)) * fn.amount END), 0) as carbs,
            COALESCE(SUM(CASE WHEN n.name = 'Fat' THEN (fd.amount_consumed / NULLIF(f.serving_size, 0)) * fn.amount END), 0) as fat
        FROM food_diaries fd
        JOIN foods f ON fd.food_id = f.id AND f.deleted_at IS NULL
        JOIN food_nutrients fn ON f.id = fn.food_id
        JOIN nutrients n ON fn.nutrient_id = n.id
        WHERE fd.user_id = $1
            AND DATE(fd.consumed_at) = $2
            AND fd.deleted_at IS NULL
    `

	summary := &domain.DailySummary{
		Entries: []domain.FoodDiary{},
	}

	err := s.db.QueryRowContext(ctx, query, userID, date).Scan(
		&summary.TotalCalories,
		&summary.TotalProtein,
		&summary.TotalCarbs,
		&summary.TotalFat,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get daily summary: %w", err)
	}

	return summary, nil
}

func (s *DiaryStore) GetEntries(ctx context.Context, userID int64, date time.Time) ([]*domain.FoodDiary, error) {
	query := `
        SELECT
            fd.id,
            fd.amount_consumed,
            fd.consumed_at,
            fd.meal_type,
            fd.created_at,
            fd.updated_at,
            f.name as food_name
        FROM food_diaries fd
        JOIN foods f ON f.id = fd.food_id
        WHERE fd.user_id = $1
          AND DATE(fd.consumed_at) = $2
          AND fd.deleted_at IS NULL
    `

	rows, err := s.db.QueryContext(ctx, query, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.FoodDiary

	for rows.Next() {
		var entry domain.FoodDiary
		err = rows.Scan(
			&entry.ID,
			&entry.AmountConsumed,
			&entry.ConsumedAt,
			&entry.MealType,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.FoodName,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

func (s *DiaryStore) GetUserEntry(ctx context.Context, userID, entryID int64) (*domain.FoodDiary, error) {
	query := `
	SELECT
		fd.user_id,
		fd.food_id,
		fd.amount_consumed,
		fd.consumed_at,
		fd.meal_type,
		f.name,
		fd.created_at,
		fd.updated_at
	FROM food_diaries fd
	JOIN foods f ON fd.food_id = f.id
	WHERE fd.id = $1 AND fd.deleted_at IS NULL AND fd.user_id = $2
	`

	diary := &domain.FoodDiary{ID: entryID}

	err := s.db.QueryRowContext(ctx, query, entryID, userID).Scan(
		&diary.UserID,
		&diary.FoodID,
		&diary.AmountConsumed,
		&diary.ConsumedAt,
		&diary.MealType,
		&diary.FoodName,
		&diary.CreatedAt,
		&diary.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return diary, nil
}

func (s *DiaryStore) GetEntry(ctx context.Context, entryID int64) (*domain.FoodDiary, error) {
	query := `
	SELECT
		fd.user_id,
		fd.food_id,
		fd.amount_consumed,
		fd.consumed_at,
		fd.meal_type,
		f.name,
		fd.created_at,
		fd.updated_at
	FROM food_diaries fd
	JOIN foods f ON fd.food_id = f.id
	WHERE fd.id = $1 AND fd.deleted_at IS NULL
	`

	diary := &domain.FoodDiary{ID: entryID}

	err := s.db.QueryRowContext(ctx, query, entryID).Scan(
		&diary.UserID,
		&diary.FoodID,
		&diary.AmountConsumed,
		&diary.ConsumedAt,
		&diary.MealType,
		&diary.FoodName,
		&diary.CreatedAt,
		&diary.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return diary, nil
}

func (s *DiaryStore) Create(ctx context.Context, entry *domain.FoodDiary) error {
	query := `
	INSERT INTO food_diaries (user_id, food_id, amount_consumed, consumed_at, meal_type)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		entry.UserID,
		entry.FoodID,
		entry.AmountConsumed,
		entry.ConsumedAt,
		entry.MealType,
	).Scan(
		&entry.ID,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *DiaryStore) Update(ctx context.Context, entry *domain.FoodDiary) error {
	query := `
	UPDATE food_diaries
		SET
			food_id = $2,
			amount_consumed = $3,
			consumed_at = $4,
			meal_type = $5,
			updated_at = NOW()
		WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query,
		entry.ID,
		entry.FoodID,
		entry.AmountConsumed,
		entry.ConsumedAt,
		entry.MealType,
	)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *DiaryStore) Delete(ctx context.Context, id int64) error {
	query := `
	UPDATE food_diaries
		SET deleted_at = NOW()
		WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
