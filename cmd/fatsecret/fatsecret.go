package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/nu7hatch/gouuid"
)

// standalone client for FatSecret development/testing
func main() {

	consumerKey := os.Getenv("FATSECRET_API_CONSUMER_KEY")
	if consumerKey == "" {
		log.Fatal("Missing consumer key")
	}

	consumerSecret := os.Getenv("FATSECRET_API_CONSUMER_SECRET")
	if consumerSecret == "" {
		log.Fatal("Missing consumer secret")
	}

	// 3-legged oauth to access a fatsecret profile: https://platform.fatsecret.com/api/Default.aspx?screen=rapitlsa

	// <HTTP Method>&<Request URL>&<Normalized Parameters>

	// calculate signature base string
	// <HTTP Method>&<Request URL>&<Normalized Parameters>

	method := "GET"
	requestTokenURL := "http://www.fatsecret.com/oauth/request_token"

	escapedRequestURL := url.QueryEscape(requestTokenURL)
	fmt.Println(escapedRequestURL)

	fmt.Println(escapedRequestURL)

	now := time.Now()
	secs := now.Unix()
	timestamp := strconv.FormatInt(secs, 10)

	params := url.Values{}
	params.Add("oauth_consumer_key", consumerKey)
	params.Add("oauth_signature_method", "HMAC-SHA1")
	params.Add("oauth_timestamp", timestamp)
	params.Add("oauth_nonce", nonce())
	params.Add("oauth_version", "1.0")
	params.Add("oauth_callback", "oob")

	//fmt.Println(params)
	//fmt.Println(params.Encode())

	encodedParams := params.Encode()
	escapedParams := url.QueryEscape(encodedParams)

	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, escapedRequestURL, escapedParams)
	fmt.Println(signatureBaseString)

	// signature
	keyStr := consumerSecret + "&"
	key := []byte(keyStr)
	hash := hmac.New(sha1.New, key)
	hash.Write([]byte(signatureBaseString))
	sig := hash.Sum(nil)


	signature := base64.StdEncoding.EncodeToString(sig)
	fmt.Println(signature)

	params.Add("oauth_signature", signature)

	req, err := http.NewRequest("GET", requestTokenURL, nil)
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

}

// generate a uuid for the oauth nonce value
func nonce() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return u.String()
}
