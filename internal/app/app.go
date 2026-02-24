package app

import (
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
)

type DBConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type Config struct {
	Db   DBConfig
	Addr string
}

type Application struct {
	Config    Config
	Store     store.Storage
	Validator *validator.Validate
}


