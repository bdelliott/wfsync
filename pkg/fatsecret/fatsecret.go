package fatsecret

import (
	"golang.org/x/oauth2"
)

// State holds state related to FatSecret API
type State struct {
	apiKey string

	Oauth2Config *oauth2.Config
}

// StateInit initializes FS state information
func StateInit(apiKey string, authCallbackURL string) *State {

	endpoint := oauth2.Endpoint{
		AuthURL:  "http://www.fatsecret.com/oauth/request_token",
		TokenURL: "http://www.fatsecret.com/oauth/authorize",
	}

	// FS uses same value for api key & secret
	cfg := &oauth2.Config{
		ClientID:     apiKey,
		ClientSecret: apiKey,
		Scopes:       []string{"user.metrics"},
		Endpoint:     endpoint,
		RedirectURL:  authCallbackURL,
	}

	withings := &State{
		Oauth2Config: cfg,
	}
	return withings
}

// TODO https://platform.fatsecret.com/api/Default.aspx?screen=rapitlsa
