package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/helper"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	store     store.Storage
	validator validator.Validate
}

func (s *UserService) GetPaginated(ctx context.Context, size, page int) ([]domain.User, error) {
	return s.store.Users.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.store.Users.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.store.Users.GetByEmail(ctx, email)
}

func (s *UserService) Create(ctx context.Context, payload domain.UserCreateInput) (*domain.UserResponse, error) {
	if err := s.validator.Struct(payload); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	if err = s.store.Users.Create(ctx, user); err != nil {
		if helper.IsDuplicateKeyError(err) {
			return nil, domain.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	userRes := domain.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return &userRes, nil
}

func (s *UserService) Update(ctx context.Context, id int64, payload domain.UserUpdateInput) (*domain.UserResponse, error) {

	if err := s.validator.Struct(payload); err != nil {
		return nil, err
	}

	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Username != nil {
		user.Username = *payload.Username
	}
	if payload.Email != nil {
		user.Email = *payload.Email
	}

	if err = s.store.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	res := domain.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return &res, nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	err := s.store.Users.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.ErrNotFound
		}

		if helper.IsForeignKeyError(err) {
			return domain.ErrCannotDelete
		}

		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
