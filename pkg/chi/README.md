# Chi Router Extension for Roamer

This package provides integration between [Roamer](https://github.com/slipros/roamer) and [Chi router](https://github.com/go-chi/chi), allowing you to easily parse URL path parameters in your HTTP handlers.

## Installation

```bash
go get -u github.com/slipros/roamer/pkg/chi@latest
```

## Features

- Seamless integration with Chi router
- Extract path parameters using the `path` struct tag
- Compatible with Roamer middleware pattern
- Type-safe path parameter parsing

## Usage

### Basic Example

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rchi "github.com/slipros/roamer/pkg/chi"
)

// UserRequest defines the structure to parse path parameters into
type UserRequest struct {
	UserID string `path:"user_id"` // Maps to the {user_id} path parameter
	Action string `path:"action"`  // Maps to the {action} path parameter
}

func main() {
	// Create Chi router
	router := chi.NewRouter()
	
	// Create Roamer instance with Chi path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rchi.NewPath(router)), // Register Chi path parser
		),
	)

	// Add middleware
	router.Use(middleware.Logger)
	
	// Route with path parameters
	router.Route("/users/{user_id}", func(r chi.Router) {
		r.With(roamer.Middleware[UserRequest](r)).Post("/{action}", handleUserAction)
	})
	
	// Start server
	http.ListenAndServe(":3000", router)
}

func handleUserAction(w http.ResponseWriter, r *http.Request) {
	// Extract parsed data from context
	var req UserRequest
	if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Use the parsed path parameters
	response := map[string]string{
		"userId": req.UserID,
		"action": req.Action,
		"status": "success",
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

### Alternative Approach Using Global Middleware

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rchi "github.com/slipros/roamer/pkg/chi"
)

type UserRequest struct {
	UserID string `path:"user_id"`
}

func main() {
	// Create Chi router
	router := chi.NewRouter()
	
	// Create Roamer instance with Chi path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rchi.NewPath(router)),
		),
	)

	// Add global middleware
	router.Use(
		middleware.Logger,
		roamer.Middleware[UserRequest](r),
	)
	
	// Define routes
	router.Post("/user/{user_id}", handleUser)
	
	// Start server
	http.ListenAndServe(":3000", router)
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
```

## How It Works

The `rchi.NewPath` function creates a path parameter parser that extracts values from Chi's route context. This adapter function is passed to `parser.NewPath` to create a parser that Roamer can use to extract path parameters from HTTP requests.

Under the hood, the adapter:

1. Uses Chi's routing context to match the request path against defined routes
2. Extracts URL parameters based on parameter names
3. Returns the parameter values to Roamer for setting struct fields

## Compatibility

This extension is compatible with:
- Go 1.18+
- Chi v5.x
- Roamer latest version

## Contributing

Contributions are welcome! If you find a bug or have an enhancement suggestion, please open an issue or submit a pull request.
