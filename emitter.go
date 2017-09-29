package fsmtwilio

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	emitable "github.com/go-carrot/fsm-emitable"
)

var (
	ACCOUNT_SID   = os.Getenv("TWILIO_ACCOUNT_SID")
	AUTH_TOKEN    = os.Getenv("TWILIO_AUTH_TOKEN")
	TWILIO_NUMBER = os.Getenv("TWILIO_NUMBER")

	ErrDeliveryFailed = errors.New("Message failed to send")
	ErrUndelivered    = errors.New("Message was not delivered")
)

type TwilioEmitter struct {
	UUID string
}

func (t *TwilioEmitter) Emit(i interface{}) error {
	client := http.Client{}
	URL := "https://api.twilio.com/2010-04-01/Accounts/" + ACCOUNT_SID + "/Messages"
	payload := url.Values{
		"To":   {t.UUID},
		"From": {TWILIO_NUMBER},
	}

	switch v := i.(type) {
	case string:
		payload.Add("Body", v)

	case emitable.Audio:
		payload.Add("MediaUrl", v.URL)

	case emitable.Image:
		payload.Add("MediaUrl", v.URL)

	case emitable.Video:
		payload.Add("MediaUrl", v.URL)

	case emitable.File:
		payload.Add("MediaUrl", v.URL)

	case emitable.QuickReply:
		body := []string{v.Message + "\n"}

		for _, reply := range v.Replies {
			body = append(body, "- '"+reply+"'\n")
		}

		payload.Add("Body", strings.Join(body, ""))

	case emitable.Typing:
		return nil

	default:
		return errors.New("TwilioEmitter cannot handle " + reflect.TypeOf(i).String())
	}

	body := payload.Encode()
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	req.SetBasicAuth(ACCOUNT_SID, AUTH_TOKEN)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	response := new(MessageSentCallback)
	err = xml.Unmarshal([]byte(resBody), &response)
	if err != nil {
		return err
	}

	if response.Message.ErrorCode != "" {
		return errors.New(response.Message.ErrorMessage)
	}

	switch response.Message.Status {
	case "failed":
		return ErrDeliveryFailed
	case "undelivered":
		return ErrUndelivered
	}

	return nil
}
