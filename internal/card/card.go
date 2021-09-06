package card

import (
	"github.com/stripe/stripe-go/v72"
	paymentIntent "github.com/stripe/stripe-go/v72/paymentIntent"
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

	pi, err := paymentIntent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

func cardErrorMessage(code stripe.ErrorCode) string {
	var msg string
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired"
	default:
		msg = "Your card was declined"
	}

	return msg
}
