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

	// protect route
	mux.Route("/admin", func(r chi.Router) {
		r.Use(app.Auth)

		r.Get("/virtual-terminal", app.VirtualTerminal)
	})

	// mux.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)
	// mux.Get("/virtual-terminal-receipt", app.VirtualTerminalReceipt)

	mux.Get("/widget/{id}", app.ChargeOnce)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Get("/receipt", app.Receipt)

	mux.Get("/plans/bronze", app.BronzePlan)
	mux.Get("/receipt/bronze", app.BronzePlanReceipt)

	// auth routes
	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)

	// serve the static file
	fileServer := http.FileServer(http.Dir("./static"))

	// handle request to /static
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
