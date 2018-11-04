package withings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/bdelliott/wfsync/pkg/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/nokiahealth"

	"context"
	"errors"
	"log"
	"net/http"
)

const code string = "code"
const csrfToken string = "taco"

// State holds state related to Withings API
type State struct {
	apiKey          string
	apiSecret       string
	authCallbackURL string

	Oauth2Config *oauth2.Config
}

// MeasurementResponse contains the json body response to a get measurement request
type MeasurementResponse struct {
	Status int
	Body   struct {
		UpdateTime    int64
		TimeZone      string
		MeasureGroups []struct {
			GroupID  int64     `json:"grpid"`
			Attrib   int       `json:"attrib"`
			Date     int64     `json:"date"`
			Category int       `json:"category"`
			DeviceID string    `json:"deviceid"`
			Measures []Measure `json:"measures"`
		} `json:"measuregrps"`
	}
}

// Measure represents a single measurement
type Measure struct {
	Value int `json:"value"`
	Type  int `json:"type"`
	Unit  int `json:"unit"`
}

// ExchangeToken swaps an authorization token for access token
func ExchangeToken(state *State, req *http.Request) (*oauth2.Token, error) {

	code := req.Form.Get(code)
	if code == "" {
		msg := "No authorization code!"
		log.Print(msg)
		return &oauth2.Token{}, errors.New(msg)
	}

	ctx := context.Background()
	token, err := state.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		msg := "Error exchanging token: " + err.Error()
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

// GetAuthorizationURL returns auth url
func GetAuthorizationURL(state *State) string {

	return state.Oauth2Config.AuthCodeURL(csrfToken, oauth2.AccessTypeOffline)
}

// GetMeasurements retrieve measurements from the Withings API
// https://developer.health.nokia.com/oauth2/#tag/measure%2Fpaths%2Fhttps%3A~1~1api.health.nokia.com~1measure%3Faction%3Dgetmeas%2Fget
func GetMeasurements(state *State, token *db.WithingsToken) (weights []db.Weight, err error) {

	const measureURL = "https://api.health.nokia.com/measure?action=getmeas"

	// url params
	const accessToken = "access_token"
	const measurementType = "meastype"
	const category = "category"
	const startdate = "startdate"
	const enddate = "enddate"
	const offsetParam = "offset"

	const weightMeasurementType = "1"
	const realMeasurement = "1"

	const minuteSeconds = 60
	const hourSeconds = minuteSeconds * 60
	const daySeconds = hourSeconds * 24
	const yearSeconds = daySeconds * 365

	params := url.Values{}
	params.Set(accessToken, token.Token.AccessToken)
	params.Set(measurementType, string(weightMeasurementType))
	params.Set(category, string(realMeasurement))

	now := time.Now()
	nowSeconds := now.Unix()

	startSeconds := nowSeconds - (yearSeconds * 10) // start 10 years ago.

	startDate := fmt.Sprint(startSeconds)
	params.Set(startdate, startDate)

	endDate := fmt.Sprint(nowSeconds)
	params.Set(enddate, endDate)

	// offset appears to not really be implemented yet, despite appearing in the docs
	offset := 0
	params.Set(offsetParam, strconv.Itoa(offset))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := state.Oauth2Config.Client(ctx, &token.Token)

	url := measureURL + "&" + params.Encode()
	log.Print(url)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var measurementResponse MeasurementResponse
	err = json.Unmarshal(body, &measurementResponse)
	if err != nil {
		log.Print("Failed to unmarshal measurement response: ", err)
		return nil, err
	}

	// convert response structure to a slice of weight values and timestamps:
	weights = make([]db.Weight, 0)

	for _, measureGroup := range measurementResponse.Body.MeasureGroups {
		measures := measureGroup.Measures
		if len(measures) != 1 {
			panic("Didn't expect more than 1 measurement in the group.")
		}

		m := measures[0]
		weightInPounds := measurementToPounds(m)
		weight := db.Weight{
			Weight:    weightInPounds,
			Timestamp: measureGroup.Date,
		}

		weights = append(weights, weight)
	}

	return weights, nil
}

func measurementToPounds(m Measure) float64 {
	// measurement is value * 10^unit --> yields kg value
	measurementKg := float64(m.Value) * math.Pow10(m.Unit)
	measurementLbs := measurementKg * 2.2
	return measurementLbs
}

// StateInit initializes withings state information
func StateInit(apiKey string, apiSecret string, authCallbackURL string) *State {

	cfg := &oauth2.Config{
		ClientID:     apiKey,
		ClientSecret: apiSecret,
		Scopes:       []string{"user.metrics"},
		Endpoint:     nokiahealth.Endpoint,
		RedirectURL:  authCallbackURL,
	}

	withings := &State{
		Oauth2Config: cfg,
	}
	return withings
}
