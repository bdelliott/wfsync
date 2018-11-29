package main

import (
	"fmt"
	"github.com/bdelliott/wfsync/pkg/fatsecret"
	"github.com/bdelliott/wfsync/pkg/oauth1"
)

// standalone client for FatSecret development/testing
func main() {

	provider := oauth1.Provider{
		RequestTokenURL: "http://www.fatsecret.com/oauth/request_token",
		CallbackURL: "oob",
		AuthorizeURL: "http://www.fatsecret.com/oauth/authorize",
		AccessTokenURL: "http://www.fatsecret.com/oauth/access_token",
		RequestURL: "http://platform.fatsecret.com/rest/server.api",
	}
	credentials := oauth1.CredentialsFromEnv("fatsecret")

	oauthClient := oauth1.Client{
		Credentials: credentials,
		Provider: provider,
	}

	// 3-legged oauth to access a fatsecret profile: https://platform.fatsecret.com/api/Default.aspx?screen=rapitlsa
	// <HTTP Method>&<Request URL>&<Normalized Parameters>

	// 1. Get a request token:
	requestToken, requestTokenSecret := oauthClient.GetRequestToken()
	fmt.Println(requestTokenSecret)

	// 2. authorize the request token:
	authorizeUrl := oauthClient.GetAuthorizeURL(requestToken)
	fmt.Println("Now visit the following URL in a browser to get a verifier code: ", authorizeUrl)


	fmt.Print("Enter verifier code from browser: ")
	var verifier int
	_, err := fmt.Scanln(&verifier)
	if err != nil {
		panic(err)
	}

	fmt.Println("Verifier code is: ", verifier)

	// 3. Get an access token:
	oauthToken, oauthTokenSecret := oauthClient.GetAccessToken(requestToken, requestTokenSecret, verifier)

	oauthClient.Token = oauthToken
	oauthClient.Secret = oauthTokenSecret


	fmt.Println("Authorization complete, now testing access to the API itself....")

	fmt.Println("user token: ", oauthClient.Token)

	// okay, now issue a test request to the API (finally)
	fsClient := fatsecret.Client{
		OAuthClient: oauthClient,
	}

	resp := fsClient.WeightsGetMonth()
	fmt.Println(resp)
}
