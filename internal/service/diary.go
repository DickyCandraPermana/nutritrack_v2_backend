package service

import (
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
)

type DiaryService struct {
	store     store.Storage
	validator validator.Validate
}
