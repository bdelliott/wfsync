package wfsync

import (
	"golang.org/x/oauth2"
	"time"
)

// Pull measurements for the user and sync to FatSecret.
func SyncUser(state *State) func (*oauth2.Token) {

	return func(nokiaToken *oauth2.Token) {

		// TODO check if FS token exists and only sync if it does.
		NokiaGetMeasurements(state.nokia, nokiaToken)
	}
}


// handle bulk pulling from the Nokia API and syncing to the FatSecret API.
func SyncWorker(state *State) {
	// TODO run this infrequently to pull user measurements in bulk
	// TODO we actually want async callbacks for daily measurements as the gold solution

	// for each nokia token, invoke synchronization of measurements for a single user:
	DBNokiaTokensGetAll(state.db, SyncUser(state))

	time.Sleep(time.Second * 30)
}


