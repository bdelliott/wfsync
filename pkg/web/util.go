package web

import (
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/bdelliott/wfsync/pkg/db"
	"github.com/bdelliott/wfsync/pkg/state"
)

// FunctionGetShortName returns the the local name of a function (without package path)
func FunctionGetShortName(fn interface{}) string {
	handlerName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	tok := strings.Split(handlerName, "/")
	shortHandlerName := tok[len(tok)-1]
	return shortHandlerName
}

// Get user information, or force a logout in the event of failure
func getUser(rw http.ResponseWriter, req *http.Request, s *state.State) (db.User, bool) {

	userId, err := getUserId(req)
	if err != nil {
		log.Fatal("Error getting user cookie: ", err)
	}

	user, exists := db.UserGet(s.DB, userId)

	if !exists {
		// user doesn't exist in the DB.  force a logout.
		http.Redirect(rw, req, "/logout", http.StatusSeeOther)
	}

	return user, exists
}
