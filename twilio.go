package fsmtwilio

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/fsm/fsm"
)

// validateRequest checks all requests for tammer
// Twlio Spec: https://www.twilio.com/docs/api/security#validating-requests
func validateRequest(r *http.Request) bool {
	r.ParseForm()

	// Identify Scheme
	scheme := r.URL.Scheme
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	// Build URL
	url := scheme + "://" + r.Host + r.RequestURI

	// Sort Params
	keys := make([]string, 0, len(r.Form))
	for k, _ := range r.Form {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build Hash
	var buffer bytes.Buffer
	buffer.WriteString(url)
	for _, k := range keys {
		buffer.WriteString(k + r.FormValue(k))
	}

	// Hash w/sha1 and base64 encode
	hash := hmac.New(sha1.New, []byte(os.Getenv("TWILIO_AUTH_TOKEN")))
	hash.Write(buffer.Bytes())
	result := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	expected := r.Header.Get("X-Twilio-Signature")

	return result == expected
}

// GetWebhook handles ...
func GetWebhook(stateMachine fsm.StateMachine, store fsm.Store) func(http.ResponseWriter, *http.Request) func(http.ResponseWritter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate Request
		if valid := validateRequest(r); !valid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		message := new(messageReceivedCallback)
		decoder := schema.NewDecoder()

		// Parse body into struct
		err := r.ParseForm()
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = decoder.Decode(message, r.PostForm)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Create Emitter
		emitter := &TwilioEmitter{
			UUID: traverser.UUID(),
		}

		// HERE
		targetutil.Step(message.From, message.Body, store, emitter, stateMachine)

		// Transition
		newState := currentState.Transition(message.Body)
		newState.EntryAction()
		traverser.SetCurrentState(newState.Slug)

		w.WriteHeader(http.StatusOK)
		return
	}
}
