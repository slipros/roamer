// Package httprouter httprouter router extensions.
package httprouter

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Path path parser for httprouter router.
func Path(r *http.Request, name string) (string, bool) {
	params := httprouter.ParamsFromContext(r.Context())
	path := params.ByName(name)
	if len(path) == 0 {
		return "", false
	}

	return path, true
}

// NewPath returns new path parser for httprouter router.
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
