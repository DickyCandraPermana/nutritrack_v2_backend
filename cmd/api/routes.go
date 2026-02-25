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
	authH *handler.AuthHandler,
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
			r.Get("/", foodH.GetFoodsHandler)
			r.Post("/", foodH.CreateFoodsHandler)

			r.Route("/{foodID}", func(r chi.Router) {
				r.Get("/", foodH.GetFoodByIdHandler)
				r.Patch("/", foodH.UpdateFoodsHandler)
				r.Delete("/", foodH.DeleteFoodsHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/", userH.GetUsersHandler)
			r.Post("/", userH.CreateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Get("/", userH.GetUserByIdHandler)
				r.Patch("/", userH.UpdateUserHandler)
				r.Delete("/", userH.DeleteUserHandler)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", userH.CreateUserHandler)
			r.Post("/login", authH.LoginHandler)
		})
	})

	return r
}
