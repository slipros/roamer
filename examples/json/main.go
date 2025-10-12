package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
)

// UserRequest demonstrates JSON parsing
type UserRequest struct {
	Name  string `json:"name" string:"trim_space"`
	Email string `json:"email" string:"trim_space,lower"`
	Age   int    `json:"age" numeric:"min=0,max=150"`
}

// UserResponse represents the API response
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	// Initialize roamer with JSON decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithFormatters(
			formatter.NewString(),
			formatter.NewNumeric(),
		),
	)

	http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		var userReq UserRequest

		if err := r.Parse(req, &userReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := UserResponse{
			ID:    "user-123",
			Name:  userReq.Name,
			Email: userReq.Email,
			Age:   userReq.Age,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/users \\")
	log.Println("  -H 'Content-Type: application/json' \\")
	log.Println("  -d '{\"name\":\" John Doe \",\"email\":\" JOHN@EXAMPLE.COM \",\"age\":25}'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
