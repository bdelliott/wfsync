package wfsync

import (
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)


func FunctionGetShortName(fn interface{}) string {
	// reflect out the package.function name of a function for logging:
	handlerName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	tok := strings.Split(handlerName, "/")
	shortHandlerName := tok[len(tok)-1]
	return shortHandlerName
}


// Get user information, or force a logout in the event of failure
func getUser(rw http.ResponseWriter, req *http.Request, state *State) (User, bool) {

	cookie, err := req.Cookie(userIdCookie)
	if err != nil {
		log.Fatal("Error getting user cookie: ", err)
	}
	userId := cookie.Value

	user, exists := DBUserGet(state.db, userId)

	if !exists {
		// user doesn't exist in the DB.  force a logout.
		http.Redirect(rw, req, "/logout", http.StatusSeeOther)
	}

	return user, exists
}


