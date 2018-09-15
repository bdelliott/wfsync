package wfsync

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/nokiahealth"

	"log"
	"net/http"
	"context"
	"errors"
)

const CODE string = "code"
const CSRF_TOKEN string = "taco"

type nokiaState struct {
	apiKey string
	apiSecret string
	authCallbackUrl string

	oauth2Config *oauth2.Config
}


// Exchange authorization token for access token
// return error string, if applicable
func NokiaExchangeToken(state *nokiaState, req *http.Request) (accessToken string, err error){

	code := req.Form.Get(CODE)
	if code == "" {
		msg := "No authorization code!"
		log.Print(msg)
		return "", errors.New(msg)
	}

	ctx := context.Background()
	token, err := state.oauth2Config.Exchange(ctx, code)
	if err != nil {
		msg := "Error exchanging token: " + err.Error();
		log.Print(msg)
		return "", err
	}

	log.Print(token)
	return "", nil
}


func NokiaGetAuthorizationUrl(state *nokiaState) string {

	return state.oauth2Config.AuthCodeURL(CSRF_TOKEN, oauth2.AccessTypeOffline)
}


func NokiaStateInit(apiKey string, apiSecret string, authCallbackUrl string) *nokiaState {

	cfg := &oauth2.Config{
		ClientID: apiKey,
		ClientSecret: apiSecret,
		Scopes: []string{"user.metrics"},
		Endpoint: nokiahealth.Endpoint,
		RedirectURL: authCallbackUrl,
	}

	nokia := &nokiaState{
		oauth2Config: cfg,
	}
	return nokia
}