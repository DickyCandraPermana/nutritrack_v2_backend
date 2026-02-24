package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MyFirstGo/internal/domain"
)

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) GetAll(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL 
    ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []domain.User

	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user) // Masukkan ke dalam slice
	}

	// 4. Cek apakah ada error saat iterasi selesai
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, password, email, created_at, updated_at
		FROM users
		WHERE email = $1
			AND deleted_at IS NULL
	`

	user := &domain.User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UserStore) Create(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (username, password, email)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Update(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users
        SET username = $2, email = $3, updated_at = NOW()
        WHERE id = $1
    `

	res, err := s.db.ExecContext(ctx, query, user.ID, user.Username, user.Email)
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

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	query := `
				UPDATE users
				SET deleted_at = NOW()
				WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query, userID)
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
