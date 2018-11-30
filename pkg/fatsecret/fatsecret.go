package fatsecret

import (
	"github.com/bdelliott/wfsync/pkg/oauth1"
	"net/url"
)


// FatSecret API client
type Client struct {
	OAuthClient oauth1.Client
}

func NewClient() Client {
	provider := oauth1.Provider{
		RequestTokenURL: "http://www.fatsecret.com/oauth/request_token",
		AuthorizeURL: "http://www.fatsecret.com/oauth/authorize",
		AccessTokenURL: "http://www.fatsecret.com/oauth/access_token",
		RequestURL: "http://platform.fatsecret.com/rest/server.api",
	}
	credentials := oauth1.CredentialsFromEnv("fatsecret")

	oauthClient := oauth1.Client{
		Credentials: credentials,
		Provider: provider,
	}

	return Client{
		OAuthClient: oauthClient,
	}
}


func (c Client) WeightsGetMonth() string {

	params := url.Values{}
	params.Add("method", "weights.get_month")
	params.Add("format", "json")

	return c.OAuthClient.Request(params)
}
