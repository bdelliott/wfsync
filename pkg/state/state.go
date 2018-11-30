package state

import (
	"database/sql"
	"github.com/bdelliott/wfsync/pkg/fatsecret"
	"github.com/gorilla/sessions"
	"os"

	"github.com/bdelliott/wfsync/pkg/withings"
)

// State is the top-level state object
type State struct {
	DB       *sql.DB
	Withings *withings.State
	FatSecret *fatsecret.State
	SessionStore *sessions.CookieStore

}

// Init initialize the main auth State data struct
func Init(db *sql.DB, withingsAuthCallbackURL string, fatSecretAuthCallbackURL string) *State {

	withingsAPIKey := os.Getenv("WITHINGS_API_KEY")
	withingsAPISecret := os.Getenv("WITHINGS_API_SECRET")

	withingsState := withings.StateInit(
		withingsAPIKey,
		withingsAPISecret,
		withingsAuthCallbackURL,
	)

	fatSecretState := fatsecret.StateInit(
		fatSecretAuthCallbackURL,
	)

	store := initSessionStore()
	state := State{
		DB:       db,
		Withings: withingsState,
		FatSecret: fatSecretState,
		SessionStore: store,
	}

	return &state
}