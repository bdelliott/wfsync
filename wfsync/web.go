package wfsync

import (
	"github.com/gorilla/context"
	"log"
	"net/http"
	"time"
)


// start a little webapp for syncing Withings (Nokia) body scale measurements to
// a FatSecret profile
func WebServe(state *State) {

	// map url paths to handler functions:

	// static assets:
	http.Handle("/js/", http.FileServer(http.Dir("./assets")))
	http.Handle("/css/", http.FileServer(http.Dir("./assets")))

	// home page:
	http.HandleFunc("/", sessionHandler(state, home))

	// login page
	http.HandleFunc("/login/", loginHandler(state))
	http.HandleFunc("/logout/", logoutHandler)

	// post-login handlers:
	http.HandleFunc("/linkNokia", sessionHandler(state, linkNokia))
	http.HandleFunc("/nokiaCallback", sessionHandler(state, nokiaCallback))

	//http.HandleFunc(authCallbackPath, authCallback(&State, &authCallbackUrl))

	// start the http service:
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler: context.ClearHandler(http.DefaultServeMux),
	}

	log.Fatal(s.ListenAndServe())
}