package wfsync

import (
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"os"
	"errors"
	"net/http"
)

const WFSYNC_STORE_KEY = "WFSYNC_STORE_KEY"

// wrap library specific store
type store struct {
	store sessions.Store
}

type session struct {
	session *sessions.Session
}


func (s *session) Get(key string, defaultValue interface{}) interface{} {

	value, ok := s.session.Values[key]

	if ok {
		return value
	} else {
		return defaultValue
	}
}

// test if key exists in the session
func (s *session) Has(key string) bool {
	_, ok := s.session.Values[key]
	return ok
}

func (s *session) Put(key string, value string) {
	s.session.Values[key] = value
}


func (s *session) Save(rw http.ResponseWriter, req *http.Request) {
	s.session.Save(req, rw)
}

func (s *store) GetSession(req *http.Request, userId string) (*session, error) {

	// start/retrieve a session withe user id for the session name
	session_, err := s.store.Get(req, userId)

	if err != nil {
		return nil, err
	} else {
		return &session{session: session_}, nil
	}
}

// retrieve or initialize a persistent store
func StoreGet() (*store, error) {

	storeKeyFilename := os.Getenv(WFSYNC_STORE_KEY)
	if storeKeyFilename == "" {
		return nil, errors.New("Missing required storage key filename.")
	}

	storeKey , err := ioutil.ReadFile(storeKeyFilename)
	if err != nil {
		log.Fatal("Failed to open key file: ", storeKeyFilename, " err: ", err.Error())
	}

	sessionStore := sessions.NewCookieStore(storeKey)
	sessionStore.MaxAge(0) // do not auto-expire

	return &store{store: sessionStore}, nil
}

