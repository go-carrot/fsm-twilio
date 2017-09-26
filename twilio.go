package fsmtwilio

import (
	"fmt"
	"net/http"
	"os"
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

// TODO Validate request integrity https://www.twilio.com/docs/api/security#validating-requests
func IncomingMessage(store fsm.Store, stateMachine fsm.StateMachine, startState string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		message := new(MessageReceivedCallback)
		decoder := schema.NewDecoder()

		// Parse body into struct
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = decoder.Decode(message, r.PostForm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
	}
}
