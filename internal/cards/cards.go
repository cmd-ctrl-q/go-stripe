package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string

	// Last four digits
	LastFour string

	// return code from stripe
	BankReturnCode string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {

	// get secret key
	stripe.Key = c.Secret

	// create payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	// Optional: Add meta data
	// params.AddMetadata("key", "value")

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

// GetPaymentMethod gets the payment method by payment intent id
func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	// create payment method
	pm, err := paymentmethod.Get(s, nil)
	if err != nil {
		return nil, err
	}
	return pm, nil
}

// RetrievePaymentIntent gets an existing payment intent by id
func (c *Card) RetrievePaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

func (c *Card) SubscribeToPlan(cust *stripe.Customer, plan, email, last4, cardType string) (string, error) {
	// get stripe customer id
	stripeCustomerID := cust.ID

	// items to subscribe customer to
	items := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(plan)},
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items:    items,
	}

	params.AddMetadata("last_four", last4)
	params.AddMetadata("card_type", cardType)
	// get payment intent from a subscription
	params.AddExpand("latest_invoice.payment_intent")
	subscription, err := sub.New(params)
	if err != nil {
		return "", err
	}

	return subscription.ID, nil
}

// CreateCustomer creates a stripe customer
func (c *Card) CreateCustomer(pm, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret

	// create customer
	customerParmas := &stripe.CustomerParams{
		PaymentMethod: stripe.String(pm),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}

	cust, err := customer.New(customerParmas)
	if err != nil {
		var msg string
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return cust, "", nil
}

// cardErrorMessage returns human readable versions of card error messages
func cardErrorMessage(code stripe.ErrorCode) string {
	var msg string
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired"
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Incorrect CVC code"
	case stripe.ErrorCodeIncorrectZip:
		msg = "Incorrect zip/postal code"
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to charge to your card"
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to charge to your card"
	default:
		msg = "Your card was declined"
	}

	return msg
}
