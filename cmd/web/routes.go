package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	// create new router
	mux := chi.NewRouter()
	mux.Use(SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("/virtual-terminal", app.VirtualTerminal)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)

	mux.Get("/widget/{id}", app.ChargeOnce)

	// serve the static file
	fileServer := http.FileServer(http.Dir("./static"))

	// handle request to /static
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
