// Package chi chi router extensions.
package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewPath returns new path parser for chi router.
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
