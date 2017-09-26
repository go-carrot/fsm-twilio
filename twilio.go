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

	"github.com/go-carrot/fsm"
	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
	"github.com/tylerb/graceful"
)

func Start(stateMachine fsm.StateMachine, startState string) {
	// Create Store
	store := &CacheStore{
		Traversers: make(map[string]fsm.Traverser, 0),
	}

	// Build Server
	srv := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    ":" + os.Getenv("PORT"),
			Handler: buildRouter(store, stateMachine, startState),
		},
	}
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func buildRouter(store fsm.Store, stateMachine fsm.StateMachine, startState string) *httprouter.Router {
	// Router
	router := &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
	}
	router.HandlerFunc(http.MethodPost, "/twilio", IncomingMessage(store, stateMachine, startState))
	return router
}

// Validation Spec: https://www.twilio.com/docs/api/security#validating-requests
func Validate(r *http.Request) bool {
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

func IncomingMessage(store fsm.Store, stateMachine fsm.StateMachine, startState string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if valid := Validate(r); !valid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		message := new(MessageReceivedCallback)
		decoder := schema.NewDecoder()

		// Parse body into struct
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = decoder.Decode(message, r.PostForm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Get Traverser
		newTraverser := false
		traverser, err := store.FetchTraverser(message.From)
		if err != nil {
			traverser, _ = store.CreateTraverser(message.From)
			traverser.SetCurrentState(startState)
			newTraverser = true
		}

		// Create Emitter
		emitter := &TwilioEmitter{
			UUID: traverser.UUID(),
		}

		// Get Current State
		currentState := stateMachine[traverser.CurrentState()](emitter, traverser)
		if newTraverser {
			currentState.EntryAction()
		}

		// Transition
		newState := currentState.Transition(message.Body)
		newState.EntryAction()
		traverser.SetCurrentState(newState.Slug)

		w.WriteHeader(http.StatusOK)
		return
	}
}
