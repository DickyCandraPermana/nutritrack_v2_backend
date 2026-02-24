package main

import (
	"net/http"
	"time"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func mountRoutes(
	appState *app.Application,
	healthH *handler.HealthHandler,
	foodH *handler.FoodHandler,
	userH *handler.UserHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", healthH.HealthCheckHandler)

		r.Route("/foods", func(r chi.Router) {
			r.Get("/", foodH.GetFoodsHandler) // Panggil dari struct handler
			r.Post("/", foodH.CreateFoodsHandler)
		})
		// r.Route("/users", ... )
	})

	return r
}
