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
	logoutTemplate = "assets/templates/logout.html"

	// cookie to identify user:
	userIdCookie = "userid"

)

// Render the home page for a logged in user
func home(rw http.ResponseWriter, req *http.Request, state *State) {

	t, err := template.ParseFiles(homeTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template %s %s", homeTemplate, err)
	}

	type HomeData struct {
		UserName string
		NokiaState string
		FatSecretState string
	}

	user, exists := getUser(rw, req, state)
	if !exists {
		return; // redirect was issued.
	}

	_, nokiaTokenExists := DBNokiaTokenGet(state.db, user)

	data := HomeData {
		UserName: user.userName,
		NokiaState: linkStr(nokiaTokenExists),
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
	} else {
		return "Not Linked"
	}
}


// Handle user login
func loginHandler(state *State) func(rw http.ResponseWriter, req *http.Request) {

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

			DBUserSave(state.db, userId, userName)

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
func linkNokia(rw http.ResponseWriter, req *http.Request, state *State) {

	nokiaAuthorizationUrl := NokiaGetAuthorizationUrl(state.nokia)
	http.Redirect(rw, req, nokiaAuthorizationUrl, http.StatusSeeOther)
}

// Logout the user
func logoutHandler(rw http.ResponseWriter, req *http.Request) {

	// delete cookie:
	cookie, err := req.Cookie(userIdCookie)
	if err != nil {
		log.Print("Failed to retrieve user id cookie: ", err)
	} else {
		log.Print("Deleting cookie")
		userId := cookie.Value

		cookie = &http.Cookie{
			Name:   userIdCookie,
			Value:  userId,
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

// To be called after user authorizes the app with the Nokia HealthMate API.
func nokiaCallback(rw http.ResponseWriter, req *http.Request, state *State) {
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

	user, exists := getUser(rw, req, state)
	if !exists {
		return; // redirect was issued.
	}

	DBNokiaTokenSave(state.db, user, token)

	http.Redirect(rw, req, "/", http.StatusFound)

}

// Wrap session related functionality common to other handlers.
func sessionHandler(state *State,
					handler func (http.ResponseWriter,
						          *http.Request,
						          *State)) func(rw http.ResponseWriter, req *http.Request) {

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

			handler(rw, req, state)
		}
	}
}


