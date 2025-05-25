// Package gorilla provides integration between the roamer package and the
// gorilla/mux router. It allows extracting path parameters from HTTP requests
// routed with gorilla/mux.
package gorilla

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Path extracts URL path parameters from HTTP requests routed with gorilla/mux.
// Uses mux.Vars() to retrieve parameters defined in route patterns.
//
// Example:
//
//	// Setup with gorilla/mux
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewPath(gorilla.Path),
//	    ),
//	)
//
//	// Request struct
//	type UserRequest struct {
//	    ID string `path:"id"` // From /users/{id} URL
//	}
func Path(r *http.Request, name string) (string, bool) {
	vars := mux.Vars(r)
	path, exists := vars[name]
	if !exists {
		return "", false
	}

	return path, true
}
