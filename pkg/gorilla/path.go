// Package gorilla mux router extensions.
package gorilla

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewPath returns new path parser for mux router.
func NewPath(r *http.Request, name string) (string, bool) {
	vars := mux.Vars(r)
	path, exists := vars[name]
	if !exists {
		return "", false
	}

	return path, true
}
