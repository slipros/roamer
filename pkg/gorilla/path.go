// Package gorilla provides integration between the roamer package and the
// gorilla/mux router. It allows extracting path parameters from HTTP requests
// routed with gorilla/mux.
package gorilla

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Path extracts a path parameter from an HTTP request routed with gorilla/mux.
// It retrieves the parameter from the request using mux.Vars.
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
//	// Create a roamer instance with the gorilla path parser
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewPath(gorilla.Path),
//	    ),
//	)
//
//	// Example struct using path parameter
//	type UserRequest struct {
//	    ID string `path:"id"` // Will be populated from the {id} parameter
//	}
//
//	// In your handler function
//	router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	    var req UserRequest
//	    if err := r.Parse(r, &req); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Use req.ID...
//	})
func Path(r *http.Request, name string) (string, bool) {
	vars := mux.Vars(r)
	path, exists := vars[name]
	if !exists {
		return "", false
	}

	return path, true
}
