package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Post("/api/payment-intent", app.GetPaymentIntent)

	mux.Get("/api/widget/{id}", app.GetWidgetByID)

	mux.Post("/api/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)

	mux.Post("/api/authenticate", app.CreateAuthToken)
	mux.Post("/api/is-authenticated", app.CheckAuthentication)
	mux.Post("/api/forgot-password", app.SendPasswordResetEmail)
	mux.Post("/api/reset-password", app.ResetPassword)

	// apply middleware to mux to protect routes
	mux.Route("/api/admin", func(r chi.Router) {
		r.Use(app.Auth)

		r.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)
		r.Post("/all-sales", app.AllSales)
		r.Post("/all-subscriptions", app.AllSubscriptions)

		r.Post("/get-sale/{id}", app.GetSale)
	})

	return mux
}
