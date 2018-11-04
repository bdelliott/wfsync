package worker

import (
	"time"

	"github.com/bdelliott/wfsync/pkg/db"
	"github.com/bdelliott/wfsync/pkg/state"
	"github.com/bdelliott/wfsync/pkg/withings"
)

// SyncUser pulls measurements for the user and syncs to FatSecret.
func SyncUser(s *state.State, withingsToken *db.WithingsToken) {

	weights, err := withings.GetMeasurements(s.Withings, withingsToken)
	if err != nil {
		panic(err)
	}

	db.WeightsSync(s.DB, withingsToken.UserID, weights)
}

// SyncWorker handles bulk pulling from the Nokia API and syncing to the FatSecret API.
func SyncWorker(s *state.State) {
	// TODO run this infrequently to pull user measurements in bulk
	// TODO we actually want async callbacks for daily measurements as the gold solution

	// for each withings token, invoke synchronization of measurements for a single user:
	withingsTokens := db.WithingsTokensGetAll(s.DB)
	for _, withingsToken := range *withingsTokens {
		SyncUser(s, &withingsToken)
	}

	time.Sleep(time.Second * 30)
}
