package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/bdelliott/wfsync/pkg/db"
	"github.com/bdelliott/wfsync/pkg/state"
	"github.com/bdelliott/wfsync/pkg/withings"
)

const (
	homeTemplate   = "assets/templates/home.html"
	loginTemplate  = "assets/templates/login.html"
	logoutTemplate = "assets/templates/logout.html"

	// cookie to identify user:
	userIDCookie = "userid"
)

// Render the home page for a logged in user
func home(rw http.ResponseWriter, req *http.Request, state *state.State) {

	t, err := template.ParseFiles(homeTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template %s %s", homeTemplate, err)
	}

	type HomeData struct {
		UserName       string
		WithingsState  string
		FatSecretState string
	}

	user, exists := getUser(rw, req, state)
	if !exists {
		return // redirect was issued.
	}

	_, withingsTokenExists := db.WithingsTokenGet(state.DB, user)

	data := HomeData{
		UserName:       user.UserName,
		WithingsState:  linkStr(withingsTokenExists),
		FatSecretState: linkStr(false),
	}
	err = t.Execute(rw, data)
	if err != nil {
		log.Fatalf("Failed to execute template %s %s", homeTemplate, err)
	}
}

func linkStr(haveToken bool) string {
	if haveToken {
		return "Linked"
	}

	return "Not Linked"
}

// Handle user login
func loginHandler(s *state.State) func(rw http.ResponseWriter, req *http.Request) {

	return func(rw http.ResponseWriter, req *http.Request) {

		if req.Method == "GET" {
			t, err := template.ParseFiles(loginTemplate)
			if err != nil {
				log.Fatalf("Failed to parse login template %s %s", loginTemplate, err)
			}

			err = t.Execute(rw, nil)
			if err != nil {
				log.Fatalf("Failed to execute login template %s %s", loginTemplate, err)
			}

		} else if req.Method == "POST" {

			req.ParseForm()
			userID := req.Form.Get("userid")
			userName := req.Form.Get("username")

			db.UserSave(s.DB, userID, userName)

			cookie := &http.Cookie{
				Name:   userIDCookie,
				Value:  userID,
				Path:   "/",
				MaxAge: 60 * 60 * 24 * 365, // a year, in seconds
			}
			http.SetCookie(rw, cookie)

			http.Redirect(rw, req, "/", http.StatusFound)
		}
	}
}

// Redirect user to the oauth login page for Withings
func linkWithings(rw http.ResponseWriter, req *http.Request, s *state.State) {

	withingsAuthorizationURL := withings.GetAuthorizationURL(s.Withings)
	http.Redirect(rw, req, withingsAuthorizationURL, http.StatusSeeOther)
}

// Logout the user
func logoutHandler(rw http.ResponseWriter, req *http.Request) {

	// delete cookie:
	cookie, err := req.Cookie(userIDCookie)
	if err != nil {
		log.Print("Failed to retrieve user id cookie: ", err)
	} else {
		log.Print("Deleting cookie")
		userID := cookie.Value

		cookie = &http.Cookie{
			Name:   userIDCookie,
			Value:  userID,
			Path:   "/",
			MaxAge: -1, // means delete now
		}
		http.SetCookie(rw, cookie)
	}

	t, err := template.ParseFiles(logoutTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template %s %s", logoutTemplate, err)
	}

	err = t.Execute(rw, nil)
	if err != nil {
		log.Fatalf("Failed to execute template %s %s", logoutTemplate, err)
	}
}

// To be called after user authorizes the app with the Withings API.
func withingsCallback(rw http.ResponseWriter, req *http.Request, s *state.State) {
	err := req.ParseForm()
	if err != nil {
		msg := fmt.Sprint("Error parsing form values", err)
		log.Print(msg)
		http.Error(rw, msg, http.StatusBadRequest)
		return
	}

	token, err := withings.ExchangeToken(s.Withings, req)
	if err != nil {
		msg := fmt.Sprint("Failed to get a Withings access token!", err)
		log.Print(msg)
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}

	log.Print("Saving token and redirecting")

	user, exists := getUser(rw, req, s)
	if !exists {
		return // redirect was issued.
	}

	db.WithingsTokenSave(s.DB, user, token)

	http.Redirect(rw, req, "/", http.StatusFound)

}

// Wrap session related functionality common to other handlers.
func sessionHandler(s *state.State,
	handler func(http.ResponseWriter,
		*http.Request,
		*state.State)) func(rw http.ResponseWriter, req *http.Request) {

	return func(rw http.ResponseWriter, req *http.Request) {

		handlerName := FunctionGetShortName(handler)

		// retrieve user id cookie - it's set in JS on the client side, so it must exist by now:
		cookie, err := req.Cookie(userIDCookie)
		if err == http.ErrNoCookie {
			// no user id cookie present, redirect to the login page.
			http.Redirect(rw, req, "/login", http.StatusSeeOther)

		} else {
			// we have a user ident, so proceed with session capable handlers
			userID := cookie.Value

			log.Printf("%s %s", handlerName, userID)

			handler(rw, req, s)
		}
	}
}
