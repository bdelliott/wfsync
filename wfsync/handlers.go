package wfsync

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const (
	homeTemplate  = "assets/templates/home.html"
	loginTemplate = "assets/templates/login.html"

	// cookie to identify user:
	userIdCookie = "userid"

	// session keys:
	sessionUserIdKey         = "userId"
	sessionUsernameKey       = "userName"

	sessionNokiaAccessToken = "nokiaAccessToken"
)

// Render the home page for a logged in user
func home(rw http.ResponseWriter, req *http.Request, state *state, session *session) {

	t, err := template.ParseFiles(homeTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template %s %s", homeTemplate, err)
	}

	type HomeData struct {
		UserName string
		NokiaState string
		FatSecretState string
	}

	userName := session.Get(sessionUsernameKey, "").(string)
	nokiaState := session.Has(sessionNokiaAccessToken)
	//fatSecretState := session.Get(sessionFatSecretStateKey, false).(bool)
	fatSecretState := false // TODO

	data := HomeData {
		UserName: userName,
		NokiaState: linkStr(nokiaState),
		FatSecretState: linkStr(fatSecretState),
	}
	err = t.Execute(rw, data)
	if err != nil {
		log.Fatalf("Failed to execute template %s %s", homeTemplate, err)
	}
}

func linkStr(stateVal bool) string {
	if stateVal {
		return "Linked"
	} else {
		return "Not Linked"
	}
}


// Handle user login
func loginHandler(state *state) func(rw http.ResponseWriter, req *http.Request) {

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
			userId := req.Form.Get("userid")
			userName := req.Form.Get("username")

			log.Print("Saving a session for user id ", userId)

			session, err := state.store.GetSession(req, userId)
			if err != nil {
				log.Fatal("Failed to get session", err)
			}

			session.Put(sessionUserIdKey, userId)
			session.Put(sessionUsernameKey, userName)
			session.Save(rw, req)

			cookie := &http.Cookie{
				Name:   userIdCookie,
				Value:  userId,
				Path:   "/",
				MaxAge: 60 * 60 * 24 * 365, // a year, in seconds
			}
			http.SetCookie(rw, cookie)

			http.Redirect(rw, req, "/", http.StatusFound)
		}
	}
}


// Redirect user to the oauth login page for Nokia HealthMate
func linkNokia(rw http.ResponseWriter, req *http.Request, state *state, session *session) {

	nokiaAuthorizationUrl := NokiaGetAuthorizationUrl(state.nokia)
	http.Redirect(rw, req, nokiaAuthorizationUrl, http.StatusSeeOther)
}

// To be called after user authorizes the app with the Nokia HealthMate API.
func nokiaCallback(rw http.ResponseWriter, req *http.Request, state *state, session *session) {
	err := req.ParseForm()
	if err != nil {
		msg := fmt.Sprint("Error parsing form values", err)
		log.Print(msg)
		http.Error(rw, msg, http.StatusBadRequest)
		return
	}

	token, err := NokiaExchangeToken(state.nokia, req)
	if err != nil {
		msg := fmt.Sprint("Failed to get a Nokia access token!", err)
		log.Print(msg)
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}

	log.Print("Saving token and redirecting")
	session.Put(sessionNokiaAccessToken, token)
	session.Save(rw, req)

	http.Redirect(rw, req, "/", http.StatusFound)

}


// Wrap session related functionality common to other handlers.
func sessionHandler(state *state,
					handler func (http.ResponseWriter,
						          *http.Request,
						          *state,
						          *session)) func(rw http.ResponseWriter, req *http.Request) {

	return func(rw http.ResponseWriter, req *http.Request) {

		handlerName := FunctionGetShortName(handler)

		// retrieve user id cookie - it's set in JS on the client side, so it must exist by now:
		cookie, err := req.Cookie(userIdCookie)
		if err == http.ErrNoCookie {
			// no user id cookie present, redirect to the login page.
			http.Redirect(rw, req, "/login", http.StatusSeeOther)

		} else {
			// we have a user ident, so proceed with session capable handlers
			userId := cookie.Value

			log.Printf("%s %s", handlerName, userId)

			session, err := state.store.GetSession(req, userId)
			if err != nil {
				msg := fmt.Sprint("Error retrieving session: ", err)
				log.Print(handlerName, msg)
				http.Error(rw, msg, http.StatusBadRequest)
			}

			handler(rw, req, state, session)
		}
	}
}


