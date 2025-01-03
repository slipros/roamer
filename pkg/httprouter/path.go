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
