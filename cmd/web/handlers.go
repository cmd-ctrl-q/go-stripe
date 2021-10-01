package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cmd-ctrl-q/go-stripe/internal/cards"
	"github.com/cmd-ctrl-q/go-stripe/internal/models"
	"github.com/cmd-ctrl-q/go-stripe/internal/urlsigner"
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

	if err := app.renderTemplate(w, r, "terminal", nil); err != nil {
		app.errorLog.Println(err)
	}
}

type TransactionData struct {
	FirstName string
	LastName  string
	Email     string

	// The ID of the PaymentIntent from the Transaction table
	PaymentIntentID string

	// The ID of the PaymentIntent from the Transaction table
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// GetTransactionData gets txn data from post and stripe
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// read posted data
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	amount, err := strconv.Atoi(paymentAmount)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// create card
	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	// get payment intent
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// get payment method
	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}

	return txnData, nil
}

// PaymentSucceeded displays the receipt page
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read posted data
	widgetID, err := strconv.Atoi(r.Form.Get("product_id"))
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// get transaction data
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new transaction
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new order
	order := models.Order{
		WidgetID:      widgetID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Write data to session to prevent customer from being charged twice if they resubmit form
	app.Session.Put(r.Context(), "receipt", txnData)

	// redirect user to another page
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

// VirtualTerminalPaymentSucceeded displays the receipt page for virtual terminal transactions
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {

	// get transaction data
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new transaction
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Write data to session to prevent customer from being charged twice if they resubmit form
	app.Session.Put(r.Context(), "receipt", txnData)

	// redirect user to another page
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

// Receipt renders a receipt page
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	// get data from session
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	// remove data from session
	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminalReceipt renders a receipt page for the virtual terminal
func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {
	// get data from session
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	// remove data from session
	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// SaveCustomer saves a custome and returns id
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return id, nil
}

func (app *application) SaveOrder(order models.Order) (int, error) {
	id, err := app.DB.InsertOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return id, nil
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

func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	// get data about bronze plan to send to client
	widget, err := app.DB.GetWidget(2)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create data types
	data := make(map[string]interface{})
	data["widget"] = widget

	// render template and send data to client
	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}

}

// LoginPage displays the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}

}

func (app *application) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	// renew session token
	app.Session.RenewToken(r.Context())

	// parse form
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// query database and compare session password with db password
	id, err := app.DB.Authenticate(email, password)
	if err != nil {
		// TODO: store error message in session to render on frontend

		// redirect
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// add user id to session
	app.Session.Put(r.Context(), "userID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	// destroy session
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	// get url
	url := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, url)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretkey),
	}

	// check if secret key is valid
	valid := signer.VerifyToken(testURL)
	if !valid {
		app.errorLog.Println("Invalid url - tampering detected")
		return
	}

	// check if expired
	expired := signer.Expired(testURL, 60)
	if expired {
		app.errorLog.Println("Link expired")
		return
	}

	data := make(map[string]interface{})
	data["email"] = r.URL.Query().Get("email")

	if err := app.renderTemplate(w, r, "reset-password", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}
