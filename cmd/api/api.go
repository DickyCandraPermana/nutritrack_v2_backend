package main

import (
	"log"
	"net/http"
	"time"

	"github.com/MyFirstGo/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type application struct {
	config    config
	store     store.Storage
	validator *validator.Validate
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	db   dbConfig
	addr string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", app.loginHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/", app.getUsers)
			r.Post("/", app.createUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Get("/", app.getUserById)
				r.Patch("/", app.updateUserHandler)
				r.Delete("/", app.deleteUserHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at server :%s", app.config.addr)

	return srv.ListenAndServe()
}
