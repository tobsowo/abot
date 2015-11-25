package auth

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/avabot/ava/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	"github.com/avabot/ava/Godeps/_workspace/src/github.com/stripe/stripe-go"
	"github.com/avabot/ava/Godeps/_workspace/src/github.com/stripe/stripe-go/charge"
	"github.com/avabot/ava/Godeps/_workspace/src/github.com/subosito/twilio"
	"github.com/avabot/ava/shared/datatypes"
	"github.com/avabot/ava/shared/mail"
	"github.com/avabot/ava/shared/sms"
)

var regexNum = regexp.MustCompile(`\d+`)

const (
	// MethodCVV will require the CVV (3-4 digit security code) for a credit
	// card on file. If the user has no credit cards on file, the user will
	// be asked for one.
	MethodCVV dt.Method = iota + 1

	// MethodZip requires the zip code associated with a credit card on
	// file. Just like MethodCVV, the user will be asked for a credit card
	// if not on file. This method is considered slightly more secure than
	// CVV, since having the physical credit card (and therefore the CVV) is
	// not enough to make a purchase.
	MethodZip

	// MethodWebCache allows a user to authenticate by clicking a link. If
	// their browser cookies have them already logged into Ava, they will be
	// authenticated. If they are not currently logged into Ava, they will
	// be asked to login. Once logged in, they will be authenticated.
	MethodWebCache

	// MethodWebLogin requires the user login to Ava on the web interface
	// using their username and password. This is the most secure option,
	// as it ensures no one has stolen the device or session token of a
	// user.
	MethodWebLogin
)

// RequestAuth ensures you're speaking to the correct user. Select the LOWEST
// level of authentication you'll allow based on a tolerance for fraud weighed
// against the convenience of the user experience. Methods are organized in
// least-secure to most-secure order. Therefore, MethodCVV will allow any
// authentication method, whereas MethodZip will only allow MethodZip and above.
// Ava will IMPROVE the quality of the authentication automatically whenever
// possible, selecting the highest authentication method for which the user has
// recently authenticated. Note that you'll never have to call RequestAuth in a
// Purchase flow. In order to drive a customer purchase, call Purchase directly,
// which will also authenticate the user.
func RequestAuth(db *sqlx.DB, tc *twilio.Client, m dt.Method, msg *dt.Msg) (
	bool, error) {
	// check last authentication date and method
	authenticated, err := msg.User.IsAuthenticated(m)
	if err != nil {
		return false, err
	}
	if authenticated {
		return true, nil
	}
	// send user confirmation text
	var t string
	switch m {
	case MethodCVV:
		t = "Please confirm a card's security code (CVC)"
	case MethodZip:
		t = "Please confirm your billing zip code"
	case MethodWebCache:
		t = "Please prove you're logged in: https://www.avabot.com/?/profile"
	case MethodWebLogin:
		if err := msg.User.DeleteSessions(db); err != nil {
			return false, err
		}
		t = "Please log in to prove it's you: https://www.avabot.com/?/login"
	}
	tx, err := db.Beginx()
	if err != nil {
		return false, err
	}
	q := `INSERT INTO authorizations (authmethod) VALUES ($1) RETURNING id`
	var aid int
	if err = tx.QueryRowx(q, m).Scan(&aid); err != nil {
		return false, err
	}
	q = `UPDATE users SET authorizationid=$1 WHERE id=$2`
	if _, err = tx.Exec(q, aid, msg.User.ID); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	if msg.Input.FlexIDType == 2 {
		if err = sms.SendMessage(tc, msg.Input.FlexID, t); err != nil {
			return false, err
		}
	} else {
		errMsg := fmt.Sprintf("unhandled flexidtype: %d",
			msg.Input.FlexIDType)
		return false, errors.New(errMsg)
	}
	return false, nil
}

// Purchase will authenticate the user and then charge a card.
func Purchase(db *sqlx.DB, tc *twilio.Client, sg *mail.Client, m dt.Method,
	msg *dt.Msg, prds []dt.Product, price uint64) error {
	if os.Getenv("AVA_ENV") == "production" {
		authenticated, err := RequestAuth(db, tc, m, msg)
		if err != nil {
			return err
		}
		if !authenticated {
			return nil
		}
	}
	desc := fmt.Sprintf("Purchase for %.2f", price)
	stripe.Key = os.Getenv("STRIPE_ACCESS_TOKEN")
	chargeParams := &stripe.ChargeParams{
		Amount:   price,
		Currency: "usd",
		Desc:     desc,
		Customer: msg.User.StripeCustomerID,
	}
	if _, err := charge.New(chargeParams); err != nil {
		return err
	}
	err := sg.SendPurchaseConfirmation(prds, price, msg.User)
	if err != nil {
		return err
	}
	return nil
}
