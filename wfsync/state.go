package wfsync

const (
	FATSECRET   = "fatsecret"
	NOKIA       = "nokia"
	USER_COOKIE = "user-id-cookie"
	USERID      = "userid"
)


type authState struct {
	nokia *nokiaState
	//nokiaRequestTokenMap map[string]*oauth.RequestToken // Nokia req token string => RequestToken
	//nokiaAccessTokenMap map[string]*oauth.AccessToken // Nokia uid => AccessToken
}

// initialize the main auth state data struct
func StateInit(nokiaApiKey string, nokiaApiSecret string, nokiaAuthCallbackUrl string) *authState {

	nokia := NokiaStateInit(
		nokiaApiKey,
		nokiaApiSecret,
		nokiaAuthCallbackUrl,
	)

	   state := authState{
	   		nokia: nokia,
	       //nokiaRequestTokenMap: make(map[string]*oauth.RequestToken),
	       //nokiaAccessTokenMap: make(map[string]*oauth.AccessToken),
	   }


	return &state
}


// User has authorized, now do something useful
/*
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
        uid := qs.Get(NOKIA_USERID)
        state.atokm[uid] = accessToken

        // save a cookie so we don't repeat the auth process a 2nd time
        c := http.Cookie{
            Name: NOKIA_USER_COOKIE,
            Value: uid,
            MaxAge: 0,
        }
        http.SetCookie(w, &c)

        // redirect back to home page
        log.Printf("Access token aquired, redirecting home...")
        http.Redirect(w, r, "/", http.StatusFound)
    }
}*/

/*
func authUser(state *authState, authCallbackUrl *string,
              w http.ResponseWriter, r *http.Request) {

    // Step 1 -
    //   Generate an oAuth token to be used for the End-User authorization call.
    //
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
*/

/*
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

*/
