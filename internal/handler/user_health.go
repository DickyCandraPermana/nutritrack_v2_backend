package handler

import (
	"net/http"

	"github.com/MyFirstGo/internal/app"
	"github.com/MyFirstGo/internal/middleware"
)

type UserHealthHandler struct {
	App *app.Application
}

func NewUserHealthHandler(app *app.Application) *UserHealthHandler {
	return &UserHealthHandler{App: app}
}

func (h UserHealthHandler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	ctx := r.Context()

	sum, err := h.App.Service.Health.GetUserHealthSummary(ctx, userID)
	if err != nil {
		h.App.ServerErrorResponse(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, sum, nil)
}
