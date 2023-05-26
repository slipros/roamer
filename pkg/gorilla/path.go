// Package gorilla mux router extensions.
package gorilla

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Path returns new path parser for mux router.
func Path(name string, r *http.Request) (string, bool) {
	vars := mux.Vars(r)
	path, exists := vars[name]
	if !exists {
		return "", false
	}

	return path, true
}
