package fsmtwilio

import (
	"os"

	"github.com/BrandonRomano/wrecker"
)

var (
	ACCOUNT_SID   = os.Getenv("TWILIO_ACCOUNT_SID")
	AUTH_TOKEN    = os.Getenv("TWILIO_AUTH_TOKEN")
	TWILIO_NUMBER = os.Getenv("TWILIO_NUMBER")
)

type TwilioEmitter struct {
	UUID string
}

func (t *TwilioEmitter) Emit(input interface{}) error {
	// TODO Validate Post was successful
	client := wrecker.New("https://api.twilio.com/2010-04-01")
	client.Post("/Accounts/"+ACCOUNT_SID+"/Messages").
		FormParam("To", t.UUID).
		FormParam("From", TWILIO_NUMBER).
		FormParam("Body", input.(string)).
		SetBasicAuth(ACCOUNT_SID, AUTH_TOKEN).
		Execute()

	return nil
}
