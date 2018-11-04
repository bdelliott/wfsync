package main

import (
	"flag"
	"log"

	"github.com/bdelliott/wfsync/pkg/db"
	"github.com/bdelliott/wfsync/pkg/state"
	"github.com/bdelliott/wfsync/pkg/web"
	"github.com/bdelliott/wfsync/pkg/worker"
)

//start a little webapp for syncing Withings (Nokia) body scale measurements
// to a FatSecret profile
func main() {

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lshortfile)

	var withingsAuthCallbackURL string

	flag.StringVar(&withingsAuthCallbackURL, "withings-auth-callback-url", "",
		"Withings Callback URL after user authorizes the app")
	flag.Parse()

	if withingsAuthCallbackURL == "" {
		log.Fatal("Missing required flag -withings-auth-callback-url")
	}

	sqlDB := db.Init()
	defer sqlDB.Close()

	s := state.Init(sqlDB, withingsAuthCallbackURL)

	go worker.SyncWorker(s)

	web.Serve(s)
}
