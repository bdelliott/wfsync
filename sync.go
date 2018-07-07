package main

import (
    "flag"
    "log"
    "net/http"
    "os"
    "github.com/mrjones/oauth"
)

const (
    USER_COOKIE = "nokia-user-id-cookie"
    USERID = "userid"
)

type authState struct {
    consumer *oauth.Consumer
    rtokm map[string]*oauth.RequestToken // req token string => RequestToken
    atokm  map[string]*oauth.AccessToken // uid => AccessToken
}


// User has authorized, now do something useful
func authCallback(state *authState, authCallbackUrl *string) func(
    w http.ResponseWriter, r *http.Request) {

    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("Handler: authcb")

        qs := r.URL.Query()

        token := qs.Get(oauth.TOKEN_PARAM)
        rtoken := state.rtokm[token]

        verifier:= qs.Get(oauth.VERIFIER_PARAM)

        // user authorized the app, we can get an access token and begin
        // interacting with the API
        consumer := state.consumer
        accessToken, err := consumer.AuthorizeToken(rtoken, verifier)

        // save mapping of user => access token
        if err != nil {
            log.Println("Error authorizing token: %v", nil)
        }

        // access tokens are permanent, so shouldn't need to refresh them
        uid := qs.Get(USERID)
        state.atokm[uid] = accessToken

        // save a cookie so we don't repeat the auth process a 2nd time
        c := http.Cookie{
            Name: USER_COOKIE,
            Value: uid,
            MaxAge: 0,
        }
        http.SetCookie(w, &c)

        // redirect back to home page
        log.Printf("Access token aquired, redirecting home...")
        http.Redirect(w, r, "/", http.StatusFound)
    }
}


// landing url, return a handler
func home(state *authState, authCallbackUrl *string) func(
    w http.ResponseWriter, r *http.Request) {

    return func(w http.ResponseWriter, r *http.Request) {

        log.Println("Handler: home")

        // if a user id cookie exists, check for an saved access token in memory
        c, err := r.Cookie(USER_COOKIE)
        if err != nil && err != http.ErrNoCookie {
           // something bad happened? 
           log.Fatalf("Unexpected error during cookie retrieval %v", err)
        }

        if err == http.ErrNoCookie {
            // do the auth flow
            authUser(state, authCallbackUrl, w, r)

        } else {
            // previously authorized
            uid := c.Value
            log.Printf("User id: %s", uid)
            accessToken := state.atokm[uid]
            log.Printf("Access token: %+v", accessToken)
        }

    }
}


func authUser(state *authState, authCallbackUrl *string,
              w http.ResponseWriter, r *http.Request) {

    /* Step 1 -
     *   Generate an oAuth token to be used for the End-User authorization call.
    */
    consumer := state.consumer
    requestToken, authorizeUrl, err := consumer.GetRequestTokenAndUrl(
        *authCallbackUrl)

    // save the request token string & associated secret for post callback
    state.rtokm[requestToken.Token] = requestToken

    if err != nil {
        log.Fatal(err)
        return
    }

    // redirect to the authorize url
    log.Println("Redirecting to the authoriztion url...")
    http.Redirect(w, r, authorizeUrl, http.StatusFound)
}


// start a little webapp for syncing Withings (Nokia) body scale measurements to
// a FatSecret profile
func main() {
    var authCallbackUrl string

    flag.StringVar(&authCallbackUrl, "auth-callback-url", "",
                   "Callback URL after user authorizes")
    flag.Parse()

    if authCallbackUrl == "" {
        log.Fatal("Missing required flag -auth-callback-url")
    }

    const authCallbackPath = "/authcb"
    authCallbackUrl += authCallbackPath // concat url path

    api_key := os.Getenv("NOKIA_API_KEY")
    api_secret := os.Getenv("NOKIA_API_SECRET")

    // https://developer.health.nokia.com/api/doc#api-OAuth_Authentication
    serviceProvider := oauth.ServiceProvider{
        RequestTokenUrl: "https://developer.health.nokia.com/account/request_token",
        AuthorizeTokenUrl: "https://developer.health.nokia.com/account/authorize",
        AccessTokenUrl: "https://developer.health.nokia.com/account/access_token",
    }

    c := oauth.NewConsumer(
        api_key,
        api_secret,
        serviceProvider,
    )

    c.Debug(false)

    state := authState{
        consumer: c,
        rtokm: make(map[string]*oauth.RequestToken),
        atokm: make(map[string]*oauth.AccessToken),
    }

    // map url paths to handler functions:
    http.HandleFunc(authCallbackPath, authCallback(&state, &authCallbackUrl))
    http.HandleFunc("/", home(&state, &authCallbackUrl))

    // start the http service:
    log.Fatal(http.ListenAndServe(":8081", nil))
}

