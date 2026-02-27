package main

import (
	"log"
	"net/http"
	"time"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/db"
	"github.com/MyFirstGo/internal/env"
	"github.com/MyFirstGo/internal/handler"
	"github.com/MyFirstGo/internal/service"
	"github.com/MyFirstGo/internal/store"
	"github.com/go-playground/validator/v10"
)

func main() {
	cfg := app.Config{
		Addr: env.GetString("ADDR", ":8080"),
		Db: app.DBConfig{
			Addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/nutritrack?sslmode=disable"),
			MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		MinioEndpoint:  env.GetString("MINIO_ADDR", "localhost:9000"),
		MinioAccessKey: env.GetString("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: env.GetString("MINIO_SECRET_KEY", "minioadminpassword"),
		MinioUseSSL:    false,
		MinioBucket:    env.GetString("MINIO_BUCKET", "avatars"),
	}

	db, err := db.New(
		cfg.Db.Addr,
		cfg.Db.MaxOpenConns,
		cfg.Db.MaxIdleConns,
		cfg.Db.MaxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	minioClient, err := app.InitMinio(cfg)
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}

	defer db.Close()
	log.Println("db connected")

	validator := validator.New()
	dbStore := store.NewStorage(db)
	minioStore := store.NewMinioStore(minioClient, "avatars")
	service := service.NewService(dbStore, *validator, minioStore)

	// 2. Init Shared App State
	appState := &app.Application{
		Config:    cfg,
		Store:     dbStore,
		Service:   service,
		Validator: validator,
		MinIO:     minioClient,
	}

	// 3. Init Handlers (Inject AppState ke sini)
	healthHandler := handler.NewHealthHandler(appState)

	authHandler := handler.NewAuthHandler(appState)

	foodHandler := handler.NewFoodHandler(appState)

	userHandler := handler.NewUserHandler(appState)

	profileHandler := handler.NewProfileHandler(appState)

	diaryHandler := handler.NewDiaryHandler(appState)

	userHealthHandler := handler.NewUserHealthHandler(appState)

	// 4. Mount Routes
	mux := mountRoutes(appState, healthHandler, authHandler, foodHandler, userHandler, profileHandler, diaryHandler, userHealthHandler)

	// 5. Run Server
	runServer(appState, mux)
}

func runServer(app *app.Application, mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at server :%s", app.Config.Addr)

	return srv.ListenAndServe()
}
