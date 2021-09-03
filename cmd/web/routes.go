package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	// create new router
	mux := chi.NewRouter()

	mux.Get("/virtual-terminal", app.VirtualTerminal)

	return mux
}
