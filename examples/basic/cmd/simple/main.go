package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
)

// CreateUserRequest defines the structure for user creation
type CreateUserRequest struct {
	// From JSON body
	Name  string `json:"name" string:"trim_space"`
	Email string `json:"email" string:"trim_space,lower"`

	// From query parameters
	Role string `query:"role" default:"user"`

	// From headers
	UserAgent string `header:"User-Agent"`
}

// UserResponse represents the API response
type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	UserAgent string `json:"user_agent"`
}

func main() {
	// Initialize roamer with needed components
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(
			parser.NewHeader(),
			parser.NewQuery(),
		),
		roamer.WithFormatters(
			formatter.NewString(),
		),
	)

	// Create HTTP handler
	http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		var userReq CreateUserRequest

		// Parse the request into the struct
		if err := r.Parse(req, &userReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create response
		response := UserResponse{
			ID:        "user-123",
			Name:      userReq.Name,
			Email:     userReq.Email,
			Role:      userReq.Role,
			UserAgent: userReq.UserAgent,
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/users?role=admin \\")
	log.Println("  -H 'Content-Type: application/json' \\")
	log.Println("  -d '{\"name\":\" John Doe \",\"email\":\"JOHN@EXAMPLE.COM\"}'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
