// Package httprouter provides integration between the roamer package and the
// julienschmidt/httprouter router. It allows extracting path parameters from
// HTTP requests routed with httprouter.
package httprouter

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Path extracts URL path parameters from HTTP requests routed with httprouter.
// Uses ParamsFromContext to retrieve parameters from the request context.
//
// Example:
//
//	// Setup with httprouter
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewPath(httprouter.Path),
//	    ),
//	)
//
//	// Request struct
//	type UserRequest struct {
//	    ID string `path:"id"` // From /users/:id URL
//	}
func Path(r *http.Request, name string) (string, bool) {
	params := httprouter.ParamsFromContext(r.Context())
	path := params.ByName(name)
	if len(path) == 0 {
		return "", false
	}

	return path, true
}

// NewPath creates a path parameter extraction function using a router instance.
// An alternative to Path when parameters aren't stored in the request context.
//
// Example:
//
//	router := httprouter.New()
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewPath(httprouter.NewPath(router)),
//	    ),
//	)
func NewPath(router *httprouter.Router) func(r *http.Request, name string) (string, bool) {
	return func(r *http.Request, name string) (string, bool) {
		if router == nil {
			return "", false
		}

		_, params, _ := router.Lookup(r.Method, r.URL.Path)
		if len(params) == 0 {
			return "", false
		}

		path := params.ByName(name)
		if len(path) == 0 {
			return "", false
		}

		return path, true
	}
}
