package main

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"log"
	"os"
)

const KEY_LENGTH = 64 // recommended value from gorilla
const KEY_FILE = "keyFile"


// generate a key for use with a gorilla session store
func main() {
	key := securecookie.GenerateRandomKey(KEY_LENGTH)

	f, err := os.Create(KEY_FILE)

	if err != nil {
		log.Fatal("Failed to create key file ", KEY_FILE)
	}

	n, err := f.Write(key)
	if err != nil {
		log.Fatal("Failed to write key to file")
	}

	fmt.Printf("Wrote %d bytes to keyfile %s\n", n, KEY_FILE)
}
