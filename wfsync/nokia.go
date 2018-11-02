package wfsync

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/nokiahealth"
	"io/ioutil"
	"net/url"
	"time"

	"context"
	"errors"
	"log"
	"net/http"
)

const CODE string = "code"
const CSRF_TOKEN string = "taco"

type nokiaState struct {
	apiKey string
	apiSecret string
	authCallbackUrl string

	Oauth2Config *oauth2.Config
}

// json body response to a get measurement request
type MeasurementResponse struct {
	Status int
	Body struct {
		UpdateTime int64
		TimeZone string
		MeasureGroups []struct {

			GroupId int64 		`json:"grpid"`
			Attrib int    		`json:"attrib"`
			Date int64    		`json:"date"`
			Category int  		`json:"category"`
			DeviceId string     `json:"deviceid"`

			Measures []struct {
				Value int 		`json:"value"`
				Type int		`json:"type"`
				Unit int		`json:"unit"`

			} `json:"measures"`

		} `json:"measuregrps"`
	}
}

// Exchange authorization token for access token
// return error string, if applicable
func NokiaExchangeToken(state *nokiaState, req *http.Request) (*oauth2.Token, error){

	code := req.Form.Get(CODE)
	if code == "" {
		msg := "No authorization code!"
		log.Print(msg)
		return &oauth2.Token{}, errors.New(msg)
	}

	ctx := context.Background()
	token, err := state.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		msg := "Error exchanging token: " + err.Error();
		log.Print(msg)
		return &oauth2.Token{}, err
	}

	// token is of type *golang.org/x/oauth2.Token
	// fields:
	// AccessToken
	// RefreshToken -- may need this to refresh
	// Expiry
	return token, nil
}


func NokiaGetAuthorizationUrl(state *nokiaState) string {

	return state.Oauth2Config.AuthCodeURL(CSRF_TOKEN, oauth2.AccessTypeOffline)
}

// retrieve measurements:
// https://developer.health.nokia.com/oauth2/#tag/measure%2Fpaths%2Fhttps%3A~1~1api.health.nokia.com~1measure%3Faction%3Dgetmeas%2Fget
func NokiaGetMeasurements(state *nokiaState, token *oauth2.Token) {
	const measureUrl = "https://api.health.nokia.com/measure?action=getmeas"

	// url params
	const accessToken = "access_token"
	const measurementType = "meastype"
	const category = "category"
	const startdate = "startdate"
	const enddate = "enddate"

	const weightMeasurementType = "1"
	const realMeasurement = "1"

	const minuteSeconds = 60
	const hourSeconds = minuteSeconds * 60
	const daySeconds = hourSeconds * 24
	const yearSeconds = daySeconds * 365

	params := url.Values{}
	params.Set(accessToken, token.AccessToken)
	params.Set(measurementType, string(weightMeasurementType))
	params.Set(category, string(realMeasurement))

	now := time.Now()
	nowSeconds := now.Unix()

	startSeconds := nowSeconds - (yearSeconds * 10) // start 10 years ago.

	startDate := fmt.Sprint(startSeconds)
	params.Set(startdate, startDate)

	endDate := fmt.Sprint(nowSeconds)
	params.Set(enddate, endDate)

	// TODO offset and lastupdate usage

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client := state.Oauth2Config.Client(ctx, token)

	url := measureUrl + "&" + params.Encode()
	log.Print(url)
	resp, err := client.Get(url)
	if err != nil {
		log.Print("Failed to fetch measurements: ", err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var measurementResponse MeasurementResponse
	err = json.Unmarshal(body, &measurementResponse)
	if err != nil {
		log.Print("Failed to unmarshal measurement response: ", err)
		return
	}

	log.Printf("%+v\n", measurementResponse)
	log.Printf("%+v\n", measurementResponse.Body)
	log.Printf("%+v\n", measurementResponse.Body.MeasureGroups)
	log.Printf("%+v\n", measurementResponse.Body.MeasureGroups[0].Measures)

	// TODO measurement is value * 10^unit --> yields kg value
	// TODO convert kg to lbs

	log.Print("whoa")
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
		Oauth2Config: cfg,
	}
	return nokia
}
