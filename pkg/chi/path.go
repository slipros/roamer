// Package chi provides integration between the roamer package and the
// go-chi/chi router. It allows extracting path parameters from HTTP requests
// routed with chi.
package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewPath creates a path parameter extraction function compatible with chi router.
// Returns a function that extracts URL path parameters from HTTP requests.
//
// Example:
//
//	// Setup with chi router
//	router := chi.NewRouter()
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewPath(chi.NewPath(router)),
//	    ),
//	)
//
//	// Example struct using path parameter
//	type UserRequest struct {
//	    ID string `path:"id"` // Will be populated from the {id} parameter
//	}
//
//	// In your handler function
//	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	    var req UserRequest
//	    if err := r.Parse(r, &req); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Use req.ID...
//	})
func NewPath(mux *chi.Mux) func(r *http.Request, name string) (string, bool) {
	return func(r *http.Request, name string) (string, bool) {
		if mux == nil {
			return "", false
		}

		rCtx := chi.NewRouteContext()
		if !mux.Match(rCtx, r.Method, r.URL.Path) {
			return "", false
		}

		path := rCtx.URLParam(name)
		if len(path) == 0 {
			return "", false
		}

		return path, true
	}
}
