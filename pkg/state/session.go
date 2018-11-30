package state

import (
	"github.com/gorilla/sessions"
	"io/ioutil"
	"os"
	"path/filepath"
)

func initSessionStore() *sessions.CookieStore {

	keyDir :=  filepath.Join(os.Getenv("HOME"), ".config", "wfsync")
	keyFile := filepath.Join(keyDir, "sessionKeyFile")

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		panic(err)
	}

	return sessions.NewCookieStore(key)
}
