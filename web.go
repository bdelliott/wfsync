package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/mrjones/oauth"
)

func getCookie(apiName string, r *http.Request) (*http.Cookie, error) {
    cookieName :=  apiName + USER_COOKIE

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

    } else {
        // previously authorized
        uid := c.Value
        log.Printf("User id: %s", uid)
        accessToken := state.nokiaAccessTokenMap[uid]
        log.Printf("Access token: %+v", accessToken)
        authorized = true

        // TODO make a trivial call to test the API here
    }

    return authorized
}

// landing url, return a handler
func home(state *authState, authCallbackUrl *string) func(
    w http.ResponseWriter, r *http.Request) {

    return func(w http.ResponseWriter, r *http.Request) {

        log.Println("Handler: home")

    }
}


func syncStatus(state *authState) func (w http.ResponseWriter, r *http.Request) {

    return func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Check sync statuses")

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
        nokiaRequestTokenMap: make(map[string]*oauth.RequestToken),
        nokiaAccessTokenMap: make(map[string]*oauth.AccessToken),
    }

    // map url paths to handler functions:
    http.Handle("/", http.FileServer(http.Dir("./assets")))
    http.HandleFunc("/syncStatus", syncStatus(&state))

    //http.HandleFunc(authCallbackPath, authCallback(&state, &authCallbackUrl))
    //#http.HandleFunc("/", home(&state, &authCallbackUrl))



    // start the http service:
    log.Fatal(http.ListenAndServe(":8080", nil))
}

