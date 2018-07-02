package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "github.com/mrjones/oauth"
)


func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Handler: authcb")

}


// landing url, return a handler
func home(authCallbackUrl *string) func(w http.ResponseWriter, r *http.Request) {

    return func(w http.ResponseWriter, r *http.Request) {

        log.Println("Handler: home")
        api_key := os.Getenv("NOKIA_API_KEY")
        api_secret := os.Getenv("NOKIA_API_SECRET")

        // https://developer.health.nokia.com/api/doc#api-OAuth_Authentication
        serviceProvider := oauth.ServiceProvider{
            RequestTokenUrl: "https://developer.health.nokia.com/account/request_token",
            AuthorizeTokenUrl: "https://developer.health.nokia.com/account/authorize",
            AccessTokenUrl: "https://developer.health.nokia.com/account/access_token",
        }

        consumer := oauth.NewConsumer(
            api_key,
            api_secret,
            serviceProvider,
        )

        consumer.Debug(false)

        /* Step 1 -
         *   Generate an oAuth token to be used for the End-User authorization call.
        */
        requestToken, authorizeUrl, err := consumer.GetRequestTokenAndUrl(*authCallbackUrl)
        fmt.Println(requestToken)

        if err != nil {
            log.Fatal(err)
            return
        }

        // redirect to the authorize url
        log.Println("Redirecting to the authoriztion url...")
        http.Redirect(w, r, authorizeUrl, http.StatusFound)

    }
}


// start a little webapp for syncing Withings (Nokia) body scale measurements to
// a FatSecret profile
func main() {
    var authCallbackUrl string

    flag.StringVar(&authCallbackUrl, "auth-callback-url", "", "Callback URL after user authorizes")
    flag.Parse()

    if authCallbackUrl == "" {
        log.Fatal("Missing required flag -auth-callback-url")
    }

    const authCallbackPath = "/authcb"
    authCallbackUrl += authCallbackPath // concat url path

    // map url paths to handler functions:
    http.HandleFunc(authCallbackPath, authCallbackHandler)
    http.HandleFunc("/", home(&authCallbackUrl))

    // start the http service:
    log.Fatal(http.ListenAndServe(":8081", nil))
}

