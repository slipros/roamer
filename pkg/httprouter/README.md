# HttpRouter Extension for Roamer

This package provides integration between [Roamer](https://github.com/slipros/roamer) and [HttpRouter](https://github.com/julienschmidt/httprouter), allowing you to easily parse URL path parameters in your HTTP handlers.

## Installation

```bash
go get -u github.com/slipros/roamer/pkg/httprouter@latest
```

## Features

- Seamless integration with HttpRouter
- Extract path parameters using the `path` struct tag
- Compatible with Roamer middleware pattern
- Type-safe path parameter parsing

## Usage

### Basic Example

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rhttprouter "github.com/slipros/roamer/pkg/httprouter"
)

// UserRequest defines the structure to parse path parameters into
type UserRequest struct {
	UserID string `path:"user_id"` // Maps to the :user_id path parameter
}

func main() {
	// Create HttpRouter
	router := httprouter.New()
	
	// Create Roamer instance with HttpRouter path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rhttprouter.Path), // Register HttpRouter path parser
		),
	)

	// Define a route with path parameters
	router.Handler(http.MethodPost, "/user/:user_id", 
		applyMiddleware(roamer.Middleware[UserRequest](r), handleUser))
	
	// Start server
	log.Println("Server starting on :3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	// Extract parsed data from context
	var req UserRequest
	if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Use the parsed path parameters
	response := map[string]string{
		"userId": req.UserID,
		"status": "success",
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// applyMiddleware chains multiple middleware functions
func applyMiddleware(middlewares ...func(http.Handler) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		
		// Apply middleware in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		
		handler.ServeHTTP(w, r)
	})
}
```

### Complete Example with Middleware Chain

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rhttprouter "github.com/slipros/roamer/pkg/httprouter"
)

// UserRequest defines the structure to parse path parameters into
type UserRequest struct {
	UserID string `path:"user_id"` // Maps to the :user_id path parameter
}

func main() {
	// Create HttpRouter
	router := httprouter.New()
	
	// Create Roamer instance with HttpRouter path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rhttprouter.Path),
		),
	)

	// Create middleware chain
	middlewareChain := Chain(
		loggingMiddleware,
		roamer.Middleware[UserRequest](r),
	)

	// Define route with middleware chain and handler
	router.Handler(http.MethodPost, "/user/:user_id", 
		middlewareChain.HandlerFunc(handleUser))
	
	// Start server
	log.Println("Server starting on :3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Respond with the parsed data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// Logging middleware example
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Chain returns a Middlewares type from a slice of middleware handlers
func Chain(middlewares ...func(http.Handler) http.Handler) Middlewares {
	return middlewares
}

// Middlewares is a slice of middleware handlers
type Middlewares []func(http.Handler) http.Handler

// Handler builds and returns a http.Handler from the chain of middlewares
func (mws Middlewares) Handler(h http.Handler) http.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// HandlerFunc builds and returns a http.Handler from the chain of middlewares
func (mws Middlewares) HandlerFunc(h http.HandlerFunc) http.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// ChainHandler is a http.Handler with support for handler composition
type ChainHandler struct {
	Endpoint    http.Handler
	chain       http.Handler
	Middlewares Middlewares
}

func (c *ChainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.chain.ServeHTTP(w, r)
}

// chain builds a http.Handler composed of middleware stack and endpoint handler
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

## How It Works

The `rhttprouter.Path` function is a path parameter parser that extracts values from HttpRouter's params object. This function is passed to `parser.NewPath` to create a parser that Roamer can use to extract path parameters from HTTP requests.

Under the hood, the adapter:

1. Uses HttpRouter's parameter extraction mechanism to get path parameters from the request context
2. Returns the parameter values to Roamer for setting struct fields

## Compatibility

This extension is compatible with:
- Go 1.18+
- HttpRouter v1.3+
- Roamer latest version

## HttpRouter vs Other Routers

HttpRouter is known for its high performance and low memory usage. It's a great choice for applications where routing performance is critical. Key differences from other routers:

- Uses a radix tree for faster routing
- Has a slightly different parameter syntax (`:param` instead of `{param}`)
- Does not support regexp in routes (for performance reasons)

## Contributing

Contributions are welcome! If you find a bug or have an enhancement suggestion, please open an issue or submit a pull request.
