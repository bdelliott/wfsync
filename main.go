package main

import (
	"flag"
	"github.com/bdelliott/withings-fatsecret-sync/wfsync"
	"log"
	"os"
)

//start a little webapp for syncing Withings (Nokia) body scale measurements
// to a FatSecret profile
func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var nokiaAuthCallbackUrl string

	flag.StringVar(&nokiaAuthCallbackUrl, "nokia-auth-callback-url", "",
		"Nokia Callback URL after user authorizes the app")
	flag.Parse()

	if nokiaAuthCallbackUrl == "" {
		log.Fatal("Missing required flag -nokia-auth-callback-url")
	}

	nokiaApiKey := os.Getenv("NOKIA_API_KEY")
	nokiaApiSecret := os.Getenv("NOKIA_API_SECRET")

	//store = sessions.NewFilesystemStore([]byte("something-very-secret"))
	authState := wfsync.StateInit(nokiaApiKey, nokiaApiSecret, nokiaAuthCallbackUrl)
	wfsync.WebServe(authState)
}
