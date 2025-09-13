package roamer_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
)

// ExampleRoamer demonstrates basic usage of the roamer package
// for parsing HTTP requests into Go structures.
func ExampleRoamer() {
	// Define a structure to hold parsed request data
	type UserRequest struct {
		ID        int    `query:"id"`
		Name      string `json:"name"`
		UserAgent string `header:"User-Agent"`
		Email     string `json:"email" string:"trim_space"`
	}

	// Create a roamer instance with parsers, decoders, and formatters
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewQuery(),  // Parse URL query parameters
			parser.NewHeader(), // Parse HTTP headers
		),
		roamer.WithDecoders(
			decoder.NewJSON(), // Decode JSON request bodies
		),
		roamer.WithFormatters(
			formatter.NewString(), // Apply string formatting
		),
	)

	// Create a sample HTTP request
	req := createSampleRequest()

	// Parse the request
	var userData UserRequest
	err := r.Parse(req, &userData)
	if err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
		return
	}

	fmt.Printf("Parsed data:\n")
	fmt.Printf("ID: %d\n", userData.ID)
	fmt.Printf("Name: %s\n", userData.Name)
	fmt.Printf("Email: %s\n", userData.Email)
	fmt.Printf("User-Agent: %s\n", userData.UserAgent)

	// Output:
	// Parsed data:
	// ID: 123
	// Name: John Doe
	// Email: john@example.com
	// User-Agent: test-agent
}

// ExampleParse demonstrates the generic Parse function for direct value retrieval.
func ExampleParse() {
	// Define a structure
	type ProductRequest struct {
		Category string  `query:"category"`
		MinPrice float64 `query:"min_price"`
		MaxPrice float64 `query:"max_price"`
	}

	// Create roamer instance
	r := roamer.NewRoamer(
		roamer.WithParsers(parser.NewQuery()),
	)

	// Create request with query parameters
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			RawQuery: "category=electronics&min_price=100.50&max_price=999.99",
		},
		Header: make(http.Header),
	}

	// Use generic Parse function
	product, err := roamer.Parse[ProductRequest](r, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Category: %s\n", product.Category)
	fmt.Printf("Price range: $%.2f - $%.2f\n", product.MinPrice, product.MaxPrice)

	// Output:
	// Category: electronics
	// Price range: $100.50 - $999.99
}

// ExampleMiddleware demonstrates using roamer as HTTP middleware.
func ExampleMiddleware() {
	type APIRequest struct {
		Action string `query:"action"`
		UserID int    `query:"user_id"`
	}

	// Create roamer instance
	r := roamer.NewRoamer(
		roamer.WithParsers(parser.NewQuery()),
	)

	// Create middleware
	middleware := roamer.Middleware[APIRequest](r)

	// Sample handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data APIRequest
		if err := roamer.ParsedDataFromContext(r.Context(), &data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("Action: %s, User ID: %d\n", data.Action, data.UserID)
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Create test request
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			RawQuery: "action=update&user_id=456",
		},
		Header: make(http.Header),
	}

	// Simulate request handling
	wrappedHandler.ServeHTTP(&mockResponseWriter{}, req)

	// Output:
	// Action: update, User ID: 456
}

// Helper functions for examples

func createSampleRequest() *http.Request {
	jsonBody := `{"name": "John Doe", "email": " john@example.com "}`
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			RawQuery: "id=123",
		},
		Header: http.Header{
			"Content-Type": {"application/json"},
			"User-Agent":   {"test-agent"},
		},
		Body:          &readCloser{bytes.NewBufferString(jsonBody)},
		ContentLength: int64(len(jsonBody)),
	}
	return req
}

type mockResponseWriter struct {
	statusCode int
}

func (m *mockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}
