package main

import (
	"net/http"
	"strconv"

	"github.com/cmd-ctrl-q/go-stripe/internal/cards"
	"github.com/go-chi/chi/v5"
)

// Home displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", nil); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminal displays the virtual terminal page
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "terminal", nil, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// PaymentSucceeded displays the receipt page
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read posted data
	cardHolder := r.Form.Get("cardholder_name")
	email := r.Form.Get("email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	// create card
	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	// get payment intent
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// get payment method
	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// create new customer

	// create new order

	// create new transaction

	// data to return with template
	data := make(map[string]interface{})
	data["cardholder"] = cardHolder
	data["email"] = email
	data["pi"] = paymentIntent
	data["pm"] = paymentMethod
	data["pa"] = paymentAmount
	data["pc"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	data["bank_return_code"] = pi.Charges.Data[0].ID

	// Write data to session and redirect user to new page
	// prevent customer from being charged twice if they resubmit form

	if err := app.renderTemplate(w, r, "succeeded", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
		return
	}
}

// ChargeOnce displays the page to buy one widge
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, err := strconv.Atoi(id)
	if err != nil {
		app.errorLog.Println("cannot convert param id to int. Possible fixes: env variable, chi/v5 import")
		return
	}

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	// serve template
	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
