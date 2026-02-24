package handler

import (
	"net/http"

	"github.com/MyFirstGo/internal/app"
)

type HealthHandler struct {
	App *app.Application
}

func NewHealthHandler(app *app.Application) *HealthHandler {
	return &HealthHandler{App: app}
}

func (h *HealthHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))

}
