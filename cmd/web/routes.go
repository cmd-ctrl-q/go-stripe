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
	mux.Get("/ws", app.WsEndPoint)

	// protect route
	mux.Route("/admin", func(r chi.Router) {
		r.Use(app.Auth)
		r.Get("/virtual-terminal", app.VirtualTerminal)
		r.Get("/all-sales", app.AllSales)
		r.Get("/all-subscriptions", app.AllSubscriptions)
		r.Get("/sales/{id}", app.ShowSale)
		r.Get("/subscriptions/{id}", app.ShowSubscription)
		r.Get("/all-users", app.AllUsers)
		r.Get("/all-users/{id}", app.OneUser)
	})

	mux.Get("/widget/{id}", app.ChargeOnce)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Get("/receipt", app.Receipt)

	mux.Get("/plans/bronze", app.BronzePlan)
	mux.Get("/receipt/bronze", app.BronzePlanReceipt)

	// auth routes
	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)
	mux.Get("/logout", app.Logout)
	mux.Get("/forgot-password", app.ForgotPassword)
	mux.Get("/reset-password", app.ShowResetPassword)

	// serve the static file
	fileServer := http.FileServer(http.Dir("./static"))

	// handle request to /static
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
