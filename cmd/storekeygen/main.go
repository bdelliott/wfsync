package main

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"log"
	"os"
	"path/filepath"
)

const KEY_LENGTH = 64 // recommended value from gorilla
const KEY_FILE = "sessionKeyFile"


// generate a key for use with a gorilla session store
func main() {

	keyDir := filepath.Join(os.Getenv("HOME"), ".config", "wfsync")
	err := os.MkdirAll(keyDir, 0700)
	if err != nil {
		panic(err)
	}

	keyFile := filepath.Join(keyDir, KEY_FILE)
	f, err := os.Create(keyFile)

	key := securecookie.GenerateRandomKey(KEY_LENGTH)

	if err != nil {
		log.Fatal("Failed to create key file ", KEY_FILE)
	}

	n, err := f.Write(key)
	if err != nil {
		log.Fatal("Failed to write key to file")
	}

	fmt.Printf("Wrote %d bytes to keyfile %s\n", n, KEY_FILE)
}
