# httprouter router extension

## Install
```go
go get -u github.com/slipros/roamer/pkg/httprouter@latest
```

## Example
```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rhttprouter "github.com/slipros/roamer/pkg/httprouter"
	"github.com/julienschmidt/httprouter"
)

type Body struct {
	UserID string `path:"user_id"`
}

func main() {
	router := httprouter.New()

	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rhttprouter.Path),
		),
	)

	handler := Chain(roamer.Middleware[Body](r)).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	router.Handler(http.MethodPost, "/user/:user_id", handler)
	
	http.ListenAndServe(":3000", router)
}

// Chain returns a Middlewares type from a slice of middleware handlers.
func Chain(middlewares ...func(http.Handler) http.Handler) Middlewares {
	return middlewares
}

type Middlewares []func(http.Handler) http.Handler

// Handler builds and returns a http.Handler from the chain of middlewares,
// with `h http.Handler` as the final handler.
func (mws Middlewares) Handler(h http.Handler) http.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// HandlerFunc builds and returns a http.Handler from the chain of middlewares,
// with `h http.Handler` as the final handler.
func (mws Middlewares) HandlerFunc(h http.HandlerFunc) http.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// ChainHandler is a http.Handler with support for handler composition and
// execution.
type ChainHandler struct {
	Endpoint    http.Handler
	chain       http.Handler
	Middlewares Middlewares
}

func (c *ChainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.chain.ServeHTTP(w, r)
}

// chain builds a http.Handler composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(middlewares []func(http.Handler) http.Handler, endpoint http.Handler) http.Handler {
	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

```