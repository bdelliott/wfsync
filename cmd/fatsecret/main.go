package main

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/mrjones/oauth"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	// standalone test client for FatSecret
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/callback/", callbackHandler)

	srv := &http.Server{
		Addr:           ":9080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        context.ClearHandler(http.DefaultServeMux),
	}

	log.Fatal(srv.ListenAndServe())
}

// called after user authorizes app
func callbackHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("callback!")
}

func homeHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("Starting oauth1 flow")

	consumerKey := os.Getenv("FATSECRET_API_CONSUMER_KEY")
	consumerSecret := os.Getenv("FATSECRET_API_CONSUMER_SECRET")
	fmt.Println(consumerKey)
	fmt.Println(consumerSecret)

	// FS uses same value for consumer secret & key
	// https://platform.fatsecret.com/api/Default.aspx?screen=rapitlsa

	c := oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "http://www.fatsecret.com/oauth/request_token",
			AuthorizeTokenUrl: "http://www.fatsecret.com/oauth/authorize",
			AccessTokenUrl:    "http://www.fatsecret.com/oauth/access_token",
			//HttpMethod: "POST",
		})

	c.Debug(true)

	requestToken, redirectUrl, err := c.GetRequestTokenAndUrl("http://localhost:9080/callback/")
	if err != nil {
		panic(err)
	}

	fmt.Println("Request token: ", requestToken)

	http.Redirect(rw, req, redirectUrl, http.StatusSeeOther)
}
