package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"time"

	"github.com/MyFirstGo/internal/domain"
	"github.com/MyFirstGo/internal/helper"
	"github.com/MyFirstGo/internal/mapper"
	"github.com/MyFirstGo/internal/store"
	"github.com/disintegration/imaging"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	store     store.Storage
	validator validator.Validate
	storage   domain.FileStorage
}

func (s *UserService) GetPaginated(ctx context.Context, size, page int) ([]*domain.User, error) {

	if page < 1 {
		page = 1
	}

	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	return s.store.Users.GetPaginated(ctx, size, offset)
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

	if payload.Weight != nil {
		user.Weight = payload.Weight
	}

	if payload.Height != nil {
		user.Height = payload.Height
	}

	if payload.DateOfBirth != nil {
		user.DateOfBirth = payload.DateOfBirth
	}

	if payload.ActivityLevel != nil {
		user.ActivityLevel = payload.ActivityLevel
	}

	if payload.Gender != nil {
		user.Gender = payload.Gender
	}

	if err = s.store.Users.Create(ctx, user); err != nil {
		if helper.IsDuplicateKeyError(err) {
			return nil, domain.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	res := mapper.UserToUserResponse(user)

	return res, nil
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

	if payload.Weight != nil {
		user.Weight = payload.Weight
	}

	if payload.Height != nil {
		user.Height = payload.Height
	}

	if payload.DateOfBirth != nil {
		user.DateOfBirth = payload.DateOfBirth
	}

	if payload.ActivityLevel != nil {
		user.ActivityLevel = payload.ActivityLevel
	}

	if payload.Gender != nil {
		user.Gender = payload.Gender
	}

	if err = s.store.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	res := mapper.UserToUserResponse(user)

	return res, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, id int64, newPassword string) (*domain.UserResponse, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)

	return mapper.UserToUserResponse(user), nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID int64, file io.Reader) (string, error) {
	src, err := imaging.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	dst := imaging.Fill(src, 200, 200, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, dst, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	fileName := fmt.Sprintf("avatars/%d-%d.jpg", userID, time.Now().Unix())

	objectName, err := s.storage.Upload(ctx, fileName, buf, int64(buf.Len()), "image/jpg")
	if err != nil {
		return "", err
	}

	if err := s.store.Users.UpdateAvatar(ctx, userID, objectName); err != nil {
		return "", err
	}

	return objectName, nil
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
