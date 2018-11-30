package web

import (
	"fmt"
	"github.com/bdelliott/wfsync/pkg/fatsecret"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"strconv"

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

	// session keys
	sessionRequestToken = "requestToken"
	sessionRequestTokenSecret = "requestTokenSecret"
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
	_, _, fatSecretTokenExists := db.FatSecretTokenGet(state.DB, user)

	data := HomeData{
		UserName:       user.UserName,
		WithingsState:  linkStr(withingsTokenExists),
		FatSecretState: linkStr(fatSecretTokenExists),
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

			err := req.ParseForm()
			if err != nil {
				panic(err)
			}

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

// Redirect user to the oauth login page for FatSecret
func linkFatSecret(rw http.ResponseWriter, req *http.Request, s *state.State) {

	client := fatsecret.NewClient()

	requestToken, requestTokenSecret := client.OAuthClient.GetRequestToken(s.FatSecret.AuthCallbackURL)
	session := getSession(s, req)

	session.Values[sessionRequestToken] = requestToken
	session.Values[sessionRequestTokenSecret] = requestTokenSecret
	err := session.Save(req, rw)
	if err != nil {
		panic(err)
	}

	userAuthorizeURL := client.OAuthClient.GetAuthorizeURL(requestToken)
	http.Redirect(rw, req, userAuthorizeURL, http.StatusSeeOther)
}

// To be called after user authorizes the app with the FatSecret API.
func fatsecretCallback(rw http.ResponseWriter, req *http.Request, s *state.State) {
	// sample url:
	// http://localhost:8080/fatsecretCallback?oauth_token=a5d5e068b1f04b158df7dbc2fc4f8a2f&oauth_verifier=7009457

	err := req.ParseForm()
	if err != nil {
		msg := fmt.Sprint("Error parsing form values", err)
		log.Print(msg)
		http.Error(rw, msg, http.StatusBadRequest)
		return
	}

	session := getSession(s, req)
	requestToken := session.Values[sessionRequestToken].(string)
	if requestToken == "" {
		log.Fatal("Request token missing from session.")
	}

	requestTokenSecret := session.Values[sessionRequestTokenSecret].(string)
	if requestTokenSecret == "" {
		log.Fatal("Request token secret missing from session.")
	}

	// confirm token passed on callback is the same as the one from the session for sanity.
	oauthToken := req.Form.Get("oauth_token")
	if oauthToken == "" {
		log.Fatal("Callback doesn't have a oauth token!?")
	}

	if requestToken != oauthToken {
		log.Fatal("Request token values differ somehow.")
	}

	verifierStr := req.Form.Get("oauth_verifier")
	if verifierStr == "" {
		log.Fatal("Missing oauth verifier")
	}
	verifier, err := strconv.Atoi(verifierStr)
	if err != nil {
		panic(err)
	}

	client := fatsecret.NewClient()
	token, secret := client.OAuthClient.GetAccessToken(requestToken, requestTokenSecret, verifier)

	log.Print("Saving token and redirecting")

	user, exists := getUser(rw, req, s)
	if !exists {
		return // redirect was issued.
	}

	// save token
	db.FatSecretTokenSave(s.DB, user, token, secret)

	http.Redirect(rw, req, "/", http.StatusFound)

}


// get or initialize a user session using the unique user id value to key into the store.
// this should only be called when the user id cookie is known to be set.
func getSession(s *state.State, req *http.Request) *sessions.Session {

	userId, _ := getUserId(req)
	session, err := s.SessionStore.Get(req, userId)
	if err != nil {
		panic(err)
	}
	return session
}


// Wrap session related functionality common to other handlers.
func sessionHandler(s *state.State,
	handler func(http.ResponseWriter,
		*http.Request,
		*state.State)) func(rw http.ResponseWriter, req *http.Request) {

	return func(rw http.ResponseWriter, req *http.Request) {

		handlerName := FunctionGetShortName(handler)

		// retrieve user id cookie - it's set in JS on the client side, so it must exist by now:
		userId, err := getUserId(req)
		if err == http.ErrNoCookie {
			// no user id cookie present, redirect to the login page.
			http.Redirect(rw, req, "/login", http.StatusSeeOther)

		} else {
			// we have a user ident, so proceed with session capable handlers
			log.Printf("%s %s", handlerName, userId)

			handler(rw, req, s)
		}
	}
}

func getUserId(req *http.Request) (userId string, err error) {

	cookie, err := req.Cookie(userIDCookie)
	if err == http.ErrNoCookie {
		// no user id cookie present, redirect to the login page.
		return "", err

	} else {
		// we have a user ident, so proceed with session capable handlers
		return cookie.Value, nil
	}
}