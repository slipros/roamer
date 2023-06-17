// Package gorilla mux router extensions.
package gorilla

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Path path parser for gorilla router.
func Path(r *http.Request, name string) (string, bool) {
	vars := mux.Vars(r)
	path, exists := vars[name]
	if !exists {
		return "", false
	}

	return path, true
}
