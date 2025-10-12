package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
)

// LoggingMiddleware demonstrates reading the request body for logging
// before it gets parsed by roamer.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body for logging
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Log the raw request body
		log.Printf("=== Incoming Request ===")
		log.Printf("Method: %s", r.Method)
		log.Printf("Path: %s", r.URL.Path)
		log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
		log.Printf("Body: %s", string(body))
		log.Printf("========================")

		// IMPORTANT: The body has been consumed!
		// Without WithPreserveBody(), the next handler won't be able to read it.
		// roamer's WithPreserveBody() option automatically restores the body.

		next.ServeHTTP(w, r)
	})
}

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	FullName string   `json:"full_name"`
	Age      int      `json:"age"`
	Tags     []string `json:"tags"`
}

// UpdateProfileRequest represents a profile update request.
type UpdateProfileRequest struct {
	Bio       string    `json:"bio"`
	Website   string    `json:"website"`
	Location  string    `json:"location"`
	BirthDate time.Time `json:"birth_date"`
}

func main() {
	// Create roamer with body preservation enabled
	// This allows the request body to be read multiple times
	rWithPreserve := roamer.NewRoamer(
		roamer.WithDecoders(
			decoder.NewJSON(),
		),
		roamer.WithPreserveBody(), // Enable body preservation
	)

	// Create roamer WITHOUT body preservation for comparison
	rWithoutPreserve := roamer.NewRoamer(
		roamer.WithDecoders(
			decoder.NewJSON(),
		),
	)

	// Example 1: Handler WITH body preservation (works correctly)
	http.Handle("/users/create", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateUserRequest

		// This works because WithPreserveBody() was used
		if err := rWithPreserve.Parse(r, &req); err != nil {
			http.Error(w, fmt.Sprintf("Parse error: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("Successfully parsed user: %s (%s)", req.Username, req.Email)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "username": "%s", "email": "%s"}`, req.Username, req.Email)
	})))

	// Example 2: Handler WITHOUT body preservation (will fail)
	http.Handle("/users/create-broken", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateUserRequest

		// This will FAIL because the body was already read by LoggingMiddleware
		// and WithPreserveBody() was NOT used
		if err := rWithoutPreserve.Parse(r, &req); err != nil {
			log.Printf("ERROR: Failed to parse (expected): %v", err)
			http.Error(w, fmt.Sprintf("Parse error (this is expected): %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true}`)
	})))

	// Example 3: Multiple reads of the same body
	http.HandleFunc("/profile/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// First, read the body to validate size
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		const maxBodySize = 10 * 1024 // 10KB
		if len(body) > maxBodySize {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		log.Printf("Body size validation: %d bytes (max: %d)", len(body), maxBodySize)

		// Now parse the body with roamer
		// This works because WithPreserveBody() restores the body after reading
		var req UpdateProfileRequest
		if err := rWithPreserve.Parse(r, &req); err != nil {
			http.Error(w, fmt.Sprintf("Parse error: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("Successfully parsed profile update: %+v", req)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "bio": "%s", "location": "%s"}`, req.Bio, req.Location)
	})

	// Example usage endpoint
	http.HandleFunc("/example", func(w http.ResponseWriter, req *http.Request) {
		examples := []string{
			"Body Preservation Examples",
			"===========================",
			"",
			"Example 1: Working endpoint (WITH body preservation)",
			"-----------------------------------------------------",
			"curl -X POST http://localhost:8080/users/create \\",
			"  -H 'Content-Type: application/json' \\",
			"  -d '{",
			"    \"username\": \"johndoe\",",
			"    \"email\": \"john@example.com\",",
			"    \"full_name\": \"John Doe\",",
			"    \"age\": 30,",
			"    \"tags\": [\"developer\", \"golang\"]",
			"  }'",
			"",
			"Expected: Success! The body is logged AND parsed correctly.",
			"",
			"Example 2: Broken endpoint (WITHOUT body preservation)",
			"-------------------------------------------------------",
			"curl -X POST http://localhost:8080/users/create-broken \\",
			"  -H 'Content-Type: application/json' \\",
			"  -d '{",
			"    \"username\": \"janedoe\",",
			"    \"email\": \"jane@example.com\",",
			"    \"full_name\": \"Jane Doe\",",
			"    \"age\": 28,",
			"    \"tags\": [\"designer\", \"ux\"]",
			"  }'",
			"",
			"Expected: ERROR! The body was consumed by logging middleware.",
			"",
			"Example 3: Multiple body reads",
			"-------------------------------",
			"curl -X POST http://localhost:8080/profile/update \\",
			"  -H 'Content-Type: application/json' \\",
			"  -d '{",
			"    \"bio\": \"Software engineer passionate about Go\",",
			"    \"website\": \"https://example.com\",",
			"    \"location\": \"San Francisco, CA\",",
			"    \"birth_date\": \"1990-01-15T00:00:00Z\"",
			"  }'",
			"",
			"Expected: Success! Body is validated for size AND parsed.",
			"",
			"Key Takeaways",
			"-------------",
			"1. Use WithPreserveBody() when you need to read the body multiple times",
			"2. Common use cases: logging, size validation, signature verification",
			"3. Without preservation, the body can only be read once",
			"4. Body preservation uses a buffer, so be mindful of memory usage",
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strings.Join(examples, "\n"))
	})

	// Start the server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("Visit http://localhost:8080/example for usage instructions")
	log.Printf("\nKey endpoints:")
	log.Printf("  POST /users/create        - Works with body preservation")
	log.Printf("  POST /users/create-broken - Fails without body preservation")
	log.Printf("  POST /profile/update      - Multiple body reads")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
