// Package chi chi router extensions.
package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewPath returns new path parser for chi router.
func NewPath(mux *chi.Mux) func(name string, r *http.Request) (string, bool) {
	return func(name string, r *http.Request) (string, bool) {
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
