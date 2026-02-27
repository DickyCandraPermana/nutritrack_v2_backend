package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func (app *Application) WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data) // Marshal dulu untuk cek error sebelum kirim status
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *Application) NotFoundResponse(w http.ResponseWriter, r *http.Request) {

	message := "the requested resource could not be found"

	app.ErrorResponse(w, r, http.StatusNotFound, message)

}

func (app *Application) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.ErrorResponse(w, r, http.StatusBadRequest, "invalid request")

}

func (app *Application) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	// Gagalkan jika ada field yang tidak dikenal di JSON
	decoder.DisallowUnknownFields()

	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	// Pastikan hanya ada satu nilai JSON di body
	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *Application) ValidationErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		// Gunakan map agar frontend gampang memetakan error ke input field
		errDetails := make(map[string]string)
		for _, e := range validationErrors {
			errDetails[e.Field()] = fmt.Sprintf("failed on the '%s' tag", e.Tag())
		}

		app.ErrorResponse(w, r, http.StatusUnprocessableEntity, errDetails)
		return
	}

	app.ServerErrorResponse(w, r, err)
}

func (app *Application) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// Structured logging: memberikan konteks tanpa mengacak-acak pesan log
	slog.Error("internal server error",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	message := "the server encountered a problem and could not process your request"
	app.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

// Update ErrorResponse untuk mendukung parameter headers nil secara default
func (app *Application) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := map[string]any{"error": message}

	if err := app.WriteJSON(w, status, env, nil); err != nil {
		slog.Error("failed to write error response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
