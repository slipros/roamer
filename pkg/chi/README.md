# chi router extension

## Install
```go
go get -u github.com/SLIpros/roamer/pkg/chi@latest
```

## Example
```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/SLIpros/roamer"
	"github.com/SLIpros/roamer/parser"
	roamerChi "github.com/SLIpros/roamer/pkg/chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Body struct {
	UserID string `path:"user_id"`
}

func main() {
	router := chi.NewRouter()

	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(roamerChi.NewPath(router)),
		),
	)

	router.Use(middleware.Logger, roamer.Middleware[Body](r))
	router.Post("/user/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		var body Body
		if err := roamer.ParsedDataFromContext(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":3000", router)
}
```