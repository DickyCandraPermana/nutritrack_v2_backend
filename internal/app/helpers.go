package app

import (
	"encoding/json"
	"log"
	"net/http"
)

func (app *Application) WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func (app *Application) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(data)
}

func (app *Application) ErrorResponse(w http.ResponseWriter, _ *http.Request, status int, message any) {
	env := map[string]any{"error": message}

	err := app.WriteJSON(w, status, env)
	if err != nil {
		log.Printf("error writing json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *Application) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ERROR: %s %s | error: %v", r.Method, r.URL.Path, err) // Log detil untuk developer

	message := "the server encountered a problem and could not process your request"
	app.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

// notFoundResponse untuk error 404
func (app *Application) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.ErrorResponse(w, r, http.StatusNotFound, message)
}

// badRequestResponse untuk error 400
func (app *Application) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}
