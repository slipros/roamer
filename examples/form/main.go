package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
)

// ContactFormRequest demonstrates URL-encoded form parsing
type ContactFormRequest struct {
	Name     string `form:"name" string:"trim_space"`
	Email    string `form:"email" string:"trim_space,lower"`
	Subject  string `form:"subject" string:"trim_space"`
	Message  string `form:"message" string:"trim_space"`
	Category string `form:"category" default:"general"`
}

// ContactResponse represents the API response
type ContactResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Subject  string `json:"subject"`
	Message  string `json:"message"`
	Category string `json:"category"`
}

func main() {
	// Initialize roamer with FormURL decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewFormURL()),
		roamer.WithFormatters(formatter.NewString()),
	)

	http.HandleFunc("/contact", func(w http.ResponseWriter, req *http.Request) {
		var contactReq ContactFormRequest

		if err := r.Parse(req, &contactReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := ContactResponse{
			ID:       "msg-789",
			Name:     contactReq.Name,
			Email:    contactReq.Email,
			Subject:  contactReq.Subject,
			Message:  contactReq.Message,
			Category: contactReq.Category,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/contact \\")
	log.Println("  -H 'Content-Type: application/x-www-form-urlencoded' \\")
	log.Println("  -d 'name=John Doe&email=JOHN@EXAMPLE.COM&subject=Help Request&message=Need assistance&category=support'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
