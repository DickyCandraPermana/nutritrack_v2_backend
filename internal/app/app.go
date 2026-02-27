package app

import (
	"github.com/MyFirstGo/internal/service"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
)

type DBConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type Config struct {
	Db             DBConfig
	Addr           string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	MinioBucket    string
}

type Application struct {
	Config    Config
	Store     store.Storage
	Service   service.Service
	Validator *validator.Validate
	MinIO     *minio.Client
	Bucket    string
}
