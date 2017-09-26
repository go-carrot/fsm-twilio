package fsmtwilio

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	client := http.Client{}
	URL := "https://api.twilio.com/2010-04-01/Accounts/" + ACCOUNT_SID + "/Messages"
	payload := url.Values{
		"To":   {t.UUID},
		"From": {TWILIO_NUMBER},
		"Body": {input.(string)},
	}
	body := payload.Encode()

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	req.SetBasicAuth(ACCOUNT_SID, AUTH_TOKEN)

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
