package oauth1

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Credentials Credentials  // client (app) credentials
	Provider Provider

	Token           string  // user access token
	Secret			string  // user token secret
}

type Credentials struct {
	consumerKey		string
	consumerSecret	string
}

type Provider struct {
	RequestTokenURL	string	// URL to get the initial request token
	AuthorizeURL	string  // URL to authorize the returned temporary token
	AccessTokenURL	string  // URL to get the final access credentials to make requests on behalf of the user

	RequestURL		string	// URL to make requests to the API once authorization is done
}

// lookup the consumer key & secret by convention from environment variables
func CredentialsFromEnv(providerName string) Credentials {

	providerName = strings.ToUpper(providerName)

	envName := fmt.Sprintf("%s_API_CONSUMER_KEY", providerName)
	consumerKey := os.Getenv(envName)
	if consumerKey == "" {
		log.Fatal("Missing consumer key")
	}

	envName = fmt.Sprintf("%s_API_CONSUMER_SECRET", providerName)
	consumerSecret := os.Getenv("FATSECRET_API_CONSUMER_SECRET")
	if consumerSecret == "" {
		log.Fatal("Missing consumer secret")
	}

	return Credentials{
		consumerKey: consumerKey,
		consumerSecret: consumerSecret,
	}
}

// get the initial request token (step 1 of authorization)
func (c Client) GetRequestToken(callbackURL string) (requestToken string, requestTokenSecret string) {

	params := url.Values{}
	params.Add("oauth_callback", callbackURL)

	resp := sendSignedRequest(c.Provider.RequestTokenURL, c.Credentials, "", "", params)

	// sample response: oauth_callback_confirmed=true&oauth_token=a0526f658e8542d5920f570b58a0ab4c&oauth_token_secret=ae0537b242e649929f3ee4e27256297a

	values, err := url.ParseQuery(resp)
	if err != nil {
		panic(err)
	}

	if values.Get("oauth_callback_confirmed") != "true" {
		log.Fatal("Callback not received")
	}

	requestToken = values.Get("oauth_token")
	if requestToken == "" {
		log.Fatal("No request token in response")
	}

	requestTokenSecret = values.Get("oauth_token_secret")
	if requestTokenSecret == "" {
		log.Fatal("No request token secret in response")
	}

	fmt.Println(requestToken)
	fmt.Println(requestTokenSecret)
	return requestToken, requestTokenSecret
}

// get the URL the client needs to visit to authorize the temporary oauth request token returned in step 1
func (c Client) GetAuthorizeURL(requestToken string) string {

	values := url.Values{}
	values.Add("oauth_token", requestToken)

	userAuthorizeURL, err := url.Parse(c.Provider.AuthorizeURL)
	if err != nil {
		panic(err)
	}
	userAuthorizeURL.RawQuery = values.Encode()

	urlStr := userAuthorizeURL.String()
	return urlStr
}

// get an access token as last step of the auth process.  user credentials are added to the client instance
func (c Client) GetAccessToken(requestToken string, requestTokenSecret string, verifier int) (oauthToken string, oauthTokenSecret string) {
	// make another signed request, but this time the key differs because it includes the token secret returned
	// with the request token

	additionalParams := url.Values{}
	additionalParams.Add("oauth_verifier", strconv.Itoa(verifier))

	resp := sendSignedRequest(c.Provider.AccessTokenURL, c.Credentials, requestToken, requestTokenSecret, additionalParams)

	values, err := url.ParseQuery(resp)
	if err != nil {
		panic(err)
	}

	oauthToken = values.Get("oauth_token")

	if oauthToken == "" {
		log.Fatal("oauth token not in response!")
	}

	oauthTokenSecret = values.Get("oauth_token_secret")

	if oauthTokenSecret == "" {
		log.Fatal("oauth token secret not in response!")
	}

	return oauthToken, oauthTokenSecret
}

// make a signed API call, params to be sent are passed in
func (c Client) Request(params url.Values) string {

	return sendSignedRequest(c.Provider.RequestURL, c.Credentials, c.Token, c.Secret, params)
}


// generate a uuid for the oauth nonce value
func nonce() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return u.String()
}

func sendSignedRequest(requestURL string, credentials Credentials, token string, secret string, additionalParams url.Values) string {
	method := "GET"

	escapedRequestURL := url.QueryEscape(requestURL)

	now := time.Now()
	secs := now.Unix()
	timestamp := strconv.FormatInt(secs, 10)

	params := url.Values{}
	params.Add("oauth_consumer_key", credentials.consumerKey)
	params.Add("oauth_signature_method", "HMAC-SHA1")
	params.Add("oauth_timestamp", timestamp)
	params.Add("oauth_nonce", nonce())
	params.Add("oauth_version", "1.0")

	if token != "" {
		params.Add("oauth_token", token)
	}

	if additionalParams != nil {
		fmt.Println("addtl params: ", additionalParams)

		for k, v := range additionalParams {
			params.Add(k, v[0]) // don't support multiple values
		}
	}

	fmt.Println("params: ", params)

	encodedParams := params.Encode()

	fmt.Println("encoded params: ", encodedParams)

	escapedParams := url.QueryEscape(encodedParams)

	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, escapedRequestURL, escapedParams)
	fmt.Println(signatureBaseString)

	// signature
	keyStr := credentials.consumerSecret + "&"
	if secret != "" {
		keyStr += secret
	}

	key := []byte(keyStr)
	hash := hmac.New(sha1.New, key)
	hash.Write([]byte(signatureBaseString))
	sig := hash.Sum(nil)


	signature := base64.StdEncoding.EncodeToString(sig)
	fmt.Println(signature)

	params.Add("oauth_signature", signature)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		panic(err)
	}

	encodedParams = params.Encode()
	req.URL.RawQuery = encodedParams

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	return string(body)
}


