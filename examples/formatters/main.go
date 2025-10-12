package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
)

// UserProfileRequest demonstrates various formatters
type UserProfileRequest struct {
	// String formatting
	Name     string `json:"name" string:"trim_space,title"`
	Username string `json:"username" string:"trim_space,lower"`
	Bio      string `json:"bio" string:"trim_space,truncate=100"`

	// Numeric constraints
	Age    int     `json:"age" numeric:"min=18,max=120"`
	Salary float64 `json:"salary" numeric:"min=0,round"`

	// Time manipulation
	BirthDate time.Time `json:"birth_date" time:"timezone=UTC,start_of_day"`

	// Slice operations
	Skills []string `json:"skills"`
	Scores []int    `json:"scores"`
}

// ProfileResponse represents the API response
type ProfileResponse struct {
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	Age       int       `json:"age"`
	Salary    float64   `json:"salary"`
	BirthDate time.Time `json:"birth_date"`
	Skills    []string  `json:"skills"`
	Scores    []int     `json:"scores"`
}

func main() {
	// Initialize roamer with formatters
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(parser.NewQuery()),
		roamer.WithFormatters(
			formatter.NewString(),
			formatter.NewNumeric(),
			formatter.NewTime(),
		),
	)

	http.HandleFunc("/profile", func(w http.ResponseWriter, req *http.Request) {
		var profileReq UserProfileRequest

		if err := r.Parse(req, &profileReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := ProfileResponse{
			Name:      profileReq.Name,
			Username:  profileReq.Username,
			Bio:       profileReq.Bio,
			Age:       profileReq.Age,
			Salary:    profileReq.Salary,
			BirthDate: profileReq.BirthDate,
			Skills:    profileReq.Skills,
			Scores:    profileReq.Scores,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/profile \\")
	log.Println("  -H 'Content-Type: application/json' \\")
	log.Println("  -d '{")
	log.Println("    \"name\":\" john DOE \",")
	log.Println("    \"username\":\" JOHN_DOE \",")
	log.Println("    \"bio\":\"A very long bio that will be truncated...\",")
	log.Println("    \"age\":25,")
	log.Println("    \"salary\":50000.75,")
	log.Println("    \"birth_date\":\"1999-05-15T10:30:00Z\",")
	log.Println("    \"skills\":[\" Go \",\" Python \",\"go\",\"RUST\",\"\"],")
	log.Println("    \"scores\":[85,92,78,95,88,82,90]")
	log.Println("  }'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
