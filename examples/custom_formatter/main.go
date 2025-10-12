package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
)

// PhoneFormatter formats phone numbers to E.164 format (+[country code][number]).
// It demonstrates how to create a custom formatter for domain-specific data.
type PhoneFormatter struct{}

// NewPhoneFormatter creates a new phone number formatter.
func NewPhoneFormatter() *PhoneFormatter {
	return &PhoneFormatter{}
}

// Tag returns the struct tag name that this formatter processes.
// This method is required by the roamer.Formatter interface.
func (f *PhoneFormatter) Tag() string {
	return "phone"
}

// Format processes the phone tag and formats the phone number.
// Supported operations:
//   - e164: Format to E.164 standard (+[country][number])
//   - strip: Strip all non-digit characters
//
// This method is required by the roamer.Formatter interface.
func (f *PhoneFormatter) Format(tag reflect.StructTag, ptr any) error {
	// Get the phone tag value
	phoneTag, ok := tag.Lookup("phone")
	if !ok {
		return nil
	}

	// Convert ptr to reflect.Value
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil
	}
	dest := val.Elem()

	// Only process string fields
	if dest.Kind() != reflect.String {
		return nil
	}

	phoneNumber := dest.String()
	if phoneNumber == "" {
		return nil
	}

	// Parse the tag to get operations
	operations := strings.Split(phoneTag, ",")

	for _, op := range operations {
		op = strings.TrimSpace(op)

		switch op {
		case "strip":
			phoneNumber = stripNonDigits(phoneNumber)

		case "e164":
			phoneNumber = formatE164(phoneNumber)

		case "":
			// Empty operation, skip
			continue

		default:
			return fmt.Errorf("unknown phone formatter operation: %s", op)
		}
	}

	// Set the formatted value back
	dest.SetString(phoneNumber)
	return nil
}

// stripNonDigits removes all non-digit characters from a string.
func stripNonDigits(s string) string {
	var result strings.Builder
	for _, ch := range s {
		if unicode.IsDigit(ch) {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// formatE164 formats a phone number to E.164 standard.
// Adds a + prefix if not present and strips non-digits.
func formatE164(s string) string {
	// Strip non-digits first
	digits := stripNonDigits(s)

	if digits == "" {
		return ""
	}

	// Add + prefix for E.164 format
	if !strings.HasPrefix(digits, "+") {
		return "+" + digits
	}

	return digits
}

// ContactRequest represents a contact form with phone number formatting.
type ContactRequest struct {
	Name  string `query:"name"`
	Email string `query:"email"`

	// Phone numbers with custom formatting
	Phone       string `query:"phone" phone:"e164"`           // Format to E.164
	AlternPhone string `query:"alt_phone" phone:"strip,e164"` // Strip then format
	Fax         string `query:"fax" phone:"strip"`            // Just strip non-digits
}

// UserRegistration demonstrates using phone formatter with JSON body.
type UserRegistration struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number" phone:"e164"` // Format phone from JSON body
	CountryCode string `json:"country_code"`
}

func main() {
	// Create a roamer instance with our custom phone formatter
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewQuery(),
		),
		roamer.WithFormatters(
			NewPhoneFormatter(),
		),
	)

	// HTTP handler that formats phone numbers from query parameters
	http.HandleFunc("/contact", func(w http.ResponseWriter, req *http.Request) {
		var contact ContactRequest

		// Parse and format the request
		if err := r.Parse(req, &contact); err != nil {
			http.Error(w, fmt.Sprintf("Parse error: %v", err), http.StatusBadRequest)
			return
		}

		// Log the formatted values
		log.Printf("Received contact request:")
		log.Printf("  Name: %s", contact.Name)
		log.Printf("  Email: %s", contact.Email)
		log.Printf("  Phone: %s (formatted to E.164)", contact.Phone)
		log.Printf("  Alt Phone: %s (formatted to E.164)", contact.AlternPhone)
		log.Printf("  Fax: %s (digits only)", contact.Fax)

		// Respond with the formatted data
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Contact information received:\n\n")
		fmt.Fprintf(w, "Name: %s\n", contact.Name)
		fmt.Fprintf(w, "Email: %s\n", contact.Email)
		fmt.Fprintf(w, "Phone (E.164): %s\n", contact.Phone)
		fmt.Fprintf(w, "Alt Phone (E.164): %s\n", contact.AlternPhone)
		fmt.Fprintf(w, "Fax (digits): %s\n", contact.Fax)
	})

	// Example endpoint with usage instructions
	http.HandleFunc("/example", func(w http.ResponseWriter, req *http.Request) {
		examples := []string{
			"Test the custom phone formatter with curl:",
			"",
			"Example 1: Basic phone formatting",
			"curl -G 'http://localhost:8080/contact' \\",
			"  --data-urlencode 'name=John Doe' \\",
			"  --data-urlencode 'email=john@example.com' \\",
			"  --data-urlencode 'phone=1234567890'",
			"",
			"Output: Phone (E.164): +1234567890",
			"",
			"Example 2: Format phone with special characters",
			"curl -G 'http://localhost:8080/contact' \\",
			"  --data-urlencode 'name=Jane Smith' \\",
			"  --data-urlencode 'email=jane@example.com' \\",
			"  --data-urlencode 'phone=+1 (555) 123-4567' \\",
			"  --data-urlencode 'alt_phone=(555) 987-6543' \\",
			"  --data-urlencode 'fax=555.111.2222'",
			"",
			"Output:",
			"  Phone (E.164): +15551234567",
			"  Alt Phone (E.164): +5559876543",
			"  Fax (digits): 5551112222",
			"",
			"Example 3: International numbers",
			"curl -G 'http://localhost:8080/contact' \\",
			"  --data-urlencode 'name=Pierre Dubois' \\",
			"  --data-urlencode 'email=pierre@example.fr' \\",
			"  --data-urlencode 'phone=33 1 42 86 82 00'",
			"",
			"Output: Phone (E.164): +33142868200",
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strings.Join(examples, "\n"))
	})

	// Demonstration endpoint showing different formatting operations
	http.HandleFunc("/demo", func(w http.ResponseWriter, req *http.Request) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"1234567890", "+1234567890"},
			{"+1234567890", "+1234567890"},
			{"(555) 123-4567", "+5551234567"},
			{"+1 (555) 123-4567", "+15551234567"},
			{"555.123.4567", "+5551234567"},
			{"1-800-FLOWERS", "+18003569377"}, // Letters to digits (if implemented)
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Phone Formatter Demo\n")
		fmt.Fprintf(w, "====================\n\n")

		for _, tc := range testCases {
			formatted := formatE164(tc.input)
			fmt.Fprintf(w, "Input:    %s\n", tc.input)
			fmt.Fprintf(w, "Output:   %s\n", formatted)
			fmt.Fprintf(w, "Expected: %s\n", tc.expected)
			if formatted == tc.expected {
				fmt.Fprintf(w, "Status:   ✓ PASS\n")
			} else {
				fmt.Fprintf(w, "Status:   ✗ FAIL\n")
			}
			fmt.Fprintf(w, "\n")
		}
	})

	// Start the server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("Visit http://localhost:8080/example for usage instructions")
	log.Printf("Visit http://localhost:8080/demo for formatter demonstration")
	log.Printf("\nQuick test: curl -G 'http://localhost:8080/contact' --data-urlencode 'name=John' --data-urlencode 'email=john@test.com' --data-urlencode 'phone=(555) 123-4567'")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
