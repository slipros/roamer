// Package httprouter provides integration between the roamer package and the
// julienschmidt/httprouter router. It allows extracting path parameters from
// HTTP requests routed with httprouter.
package httprouter

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Path extracts a path parameter from an HTTP request routed with httprouter.
// It retrieves the parameter from the request context using httprouter.ParamsFromContext.
//
// Parameters:
//   - r: The HTTP request containing the path parameters.
//   - name: The name of the path parameter to extract.
//
// Returns:
//   - string: The value of the path parameter.
//   - bool: Whether the parameter was found.
//
// Example:
//
//	// Create a roamer instance with the httprouter path parser
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewPath(httprouter.Path),
//	    ),
//	)
//
//	// Example struct using path parameter
//	type UserRequest struct {
//	    ID string `path:"id"` // Will be populated from the :id parameter
//	}
//
//	// In your handler function
//	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	    var req UserRequest
//	    if err := r.Parse(r, &req); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Use req.ID...
//	})
func Path(r *http.Request, name string) (string, bool) {
	params := httprouter.ParamsFromContext(r.Context())
	path := params.ByName(name)
	if len(path) == 0 {
		return "", false
	}

	return path, true
}

// NewPath creates a path parameter extraction function that uses the provided
// httprouter.Router instance to look up path parameters. This is an alternative
// to the Path function that doesn't require parameters to be stored in the request context.
//
// This function is less commonly used than Path, but can be useful in situations
// where the router is not configured to store parameters in the context.
//
// Parameters:
//   - router: The httprouter.Router instance to use for parameter lookup.
//
// Returns:
//   - A function that extracts path parameters from HTTP requests.
//
// Example:
//
//	// Create a router
//	router := httprouter.New()
//
//	// Create a roamer instance with the httprouter path parser
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
