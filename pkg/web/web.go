package web

import (
	"log"
	"net/http"
	"time"

	"github.com/bdelliott/wfsync/pkg/state"
	"github.com/gorilla/context"
)

// Serve starts a little webapp for syncing Withings body scale measurements to
// a FatSecret profile
func Serve(s *state.State) {

	// map url paths to handler functions:

	// static assets:
	http.Handle("/js/", http.FileServer(http.Dir("./assets")))
	http.Handle("/css/", http.FileServer(http.Dir("./assets")))

	// home page:
	http.HandleFunc("/", sessionHandler(s, home))

	// login page
	http.HandleFunc("/login/", loginHandler(s))
	http.HandleFunc("/logout/", logoutHandler)

	// post-login handlers:
	http.HandleFunc("/linkWithings", sessionHandler(s, linkWithings))
	http.HandleFunc("/withingsCallback", sessionHandler(s, withingsCallback))

	//http.HandleFunc(authCallbackPath, authCallback(&State, &authCallbackUrl))

	// start the http service:
	srv := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        context.ClearHandler(http.DefaultServeMux),
	}

	log.Fatal(srv.ListenAndServe())
}
