# Gorilla Mux Router Extension for Roamer

This package provides integration between [Roamer](https://github.com/slipros/roamer) and [Gorilla Mux](https://github.com/gorilla/mux), allowing you to easily parse URL path parameters in your HTTP handlers.

## Installation

```bash
go get -u github.com/slipros/roamer/pkg/gorilla@latest
```

## Features

- Seamless integration with Gorilla Mux router
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

	"github.com/gorilla/mux"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rgorilla "github.com/slipros/roamer/pkg/gorilla"
)

// UserRequest defines the structure to parse path parameters into
type UserRequest struct {
	UserID   string `path:"user_id"`   // Maps to the {user_id} path parameter
	Resource string `path:"resource"`  // Maps to the {resource} path parameter
}

func main() {
	// Create Gorilla Mux router
	router := mux.NewRouter()
	
	// Create Roamer instance with Gorilla path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rgorilla.Path), // Register Gorilla path parser
		),
	)

	// Define a route with path parameters and middleware
	router.Handle("/users/{user_id}/{resource}", 
		roamer.Middleware[UserRequest](r)(http.HandlerFunc(handleUserResource))).
		Methods(http.MethodGet)
	
	// Start server
	log.Println("Server starting on :3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleUserResource(w http.ResponseWriter, r *http.Request) {
	// Extract parsed data from context
	var req UserRequest
	if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Use the parsed path parameters
	response := map[string]string{
		"userId": req.UserID,
		"resource": req.Resource,
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
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rgorilla "github.com/slipros/roamer/pkg/gorilla"
)

type UserRequest struct {
	UserID string `path:"user_id"`
}

func main() {
	// Create Gorilla Mux router
	router := mux.NewRouter()
	
	// Create Roamer instance with Gorilla path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewPath(rgorilla.Path),
		),
	)

	// Add global middleware
	router.Use(roamer.Middleware[UserRequest](r))
	
	// Define routes
	router.HandleFunc("/user/{user_id}", handleUser).Methods(http.MethodPost)
	
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
```

## How It Works

The `rgorilla.Path` function is a path parameter parser that extracts values from Gorilla Mux's route variables. This function is passed to `parser.NewPath` to create a parser that Roamer can use to extract path parameters from HTTP requests.

Under the hood, the adapter:

1. Uses Gorilla Mux's `mux.Vars()` function to extract path variables from the request
2. Returns the parameter values to Roamer for setting struct fields

## Compatibility

This extension is compatible with:
- Go 1.18+
- Gorilla Mux v1.8+
- Roamer latest version

## Advantages over Manual Extraction

Using Roamer with Gorilla Mux offers several advantages over manually extracting path parameters:

1. **Type Safety**: Path parameters are automatically converted to the correct type
2. **Centralized Definition**: Define your request structure in one place
3. **Validation**: Combine with formatters for post-processing and validation
4. **Code Reduction**: Less boilerplate code in your handlers

## Contributing

Contributions are welcome! If you find a bug or have an enhancement suggestion, please open an issue or submit a pull request.
