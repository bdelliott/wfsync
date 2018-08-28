package wfsync

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func getCookie(apiName string, r *http.Request) (*http.Cookie, error) {
	cookieName := apiName + USER_COOKIE

	c, err := r.Cookie(cookieName)
	if err != nil && err != http.ErrNoCookie {
		// something bad happened?
		log.Fatalf("Unexpected error during cookie retrieval %v", err)
	}

	return c, err
}

// check for prior FatSecret API authorization
func getFatSecretStatus(state *authState, r *http.Request) bool {
	authorized := false
	return authorized
}

// check for prior Nokia API authorization
func getNokiaStatus(state *authState, r *http.Request) bool {
	authorized := false

	c, err := getCookie(NOKIA, r)
	if err == http.ErrNoCookie {
		// TODO insert auth elsewhere after user action to do it.
		// do the auth flow
		//authUser(state, authCallbackUrl, w, r)
		log.Printf("No Nokia cookie present")

	} else {
		// previously authorized
		uid := c.Value
		log.Printf("User id: %s", uid)
		//accessToken := state.nokiaAccessTokenMap[uid]
		//log.Printf("Access token: %+v", accessToken)
		authorized = true

		// TODO make a trivial call to test the API here
	}

	return authorized
}

// To be called after user authorizes the app with the Nokia HealthMate API.
func nokiaCallback(state *authState) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			msg := fmt.Sprintf("Error parsing form values", err)
			logRequestInfo(req, msg)
			http.Error(rw, msg, http.StatusBadRequest)
		} else {
			logRequestInfo(req, "nokiaCallback")
			token, err := NokiaExchangeToken(state.nokia, req)
			if err != nil {
				log.Fatal("boom")
			}
			log.Print(token)

		}

	}
}

// Redirect user to the
func syncNokia(state *authState) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		logRequestInfo(req, "syncNokia")

        nokiaAuthorizationUrl := NokiaGetAuthorizationUrl(state.nokia)

	    m := make(map[string]string)
	    m["url"] = nokiaAuthorizationUrl

	    buf, _ := json.Marshal(m)
	    fmt.Fprintln(rw, string(buf))
	}
}

func syncStatus(state *authState) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")

		logRequestInfo(r, "Check sync statuses")

		// check if the user previously authorized access to the Nokia/Withings API
		nokiaStatus := getNokiaStatus(state, r)
		log.Printf("Nokia authorized?: %v", nokiaStatus)

		// check if the user previously authorized access to the FatSecret API
		fsStatus := getFatSecretStatus(state, r)

		m := make(map[string]bool)
		m[NOKIA] = nokiaStatus
		m[FATSECRET] = fsStatus

		buf, _ := json.Marshal(m)
		fmt.Fprintln(w, string(buf))
	}
}

// log some common http request info and a message
func logRequestInfo(req *http.Request, msg string) {
	log.Printf("%s %s :: %s", req.Method, req.URL, msg)
}

// start a little webapp for syncing Withings (Nokia) body scale measurements to
// a FatSecret profile
func WebServe(state *authState) {

	// map url paths to handler functions:
	http.Handle("/", http.FileServer(http.Dir("./assets")))
	http.HandleFunc("/nokiaCallback", nokiaCallback(state));
	http.HandleFunc("/syncNokia", syncNokia(state))
	http.HandleFunc("/syncStatus", syncStatus(state))

	//http.HandleFunc(authCallbackPath, authCallback(&state, &authCallbackUrl))

	// start the http service:
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
