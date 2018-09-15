package wfsync

import (
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


