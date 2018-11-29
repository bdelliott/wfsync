package fatsecret

import (
	"github.com/bdelliott/wfsync/pkg/oauth1"
	"net/url"
)

// FatSecret API client
type Client struct {
	OAuthClient oauth1.Client
}


func (c Client) WeightsGetMonth() string {

	params := url.Values{}
	params.Add("method", "weights.get_month")
	params.Add("format", "json")

	return c.OAuthClient.Request(params)
}
