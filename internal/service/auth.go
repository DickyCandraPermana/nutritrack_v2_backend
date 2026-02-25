package service

import (
	"context"
	"log"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/helper"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	store     store.Storage
	validator validator.Validate
}

func (s *AuthService) Login(ctx context.Context, payload domain.UserLoginInput) (*domain.LoginResponse, error) {
	if err := s.validator.Struct(payload); err != nil {
		return nil, err
	}

	user, err := s.store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		log.Printf("Failed login attempt for user ID: %d", user.ID)
		return nil, domain.ErrInvalidCredentials
	}

	token, err := helper.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate token for user %d: %v", user.ID, err)
		return nil, err
	}

	return &domain.LoginResponse{
		Token: token,
		Type:  "Bearer",
	}, nil
}
