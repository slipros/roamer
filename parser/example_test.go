package parser_test

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
)

// ExampleNewQuery demonstrates creating and using a query parameter parser.
func ExampleNewQuery() {
	// Define a structure with query parameter tags
	type SearchRequest struct {
		Query    string   `query:"q"`
		Category string   `query:"category"`
		Tags     []string `query:"tags"` // Will split comma-separated values
		Page     int      `query:"page"`
		Limit    int      `query:"limit"`
	}

	// Create a query parser
	queryParser := parser.NewQuery()

	// Create a roamer instance with the query parser
	r := roamer.NewRoamer(
		roamer.WithParsers(queryParser),
	)

	// Create a request with query parameters
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			RawQuery: "q=golang&category=programming&tags=web,api,http&page=1&limit=10",
		},
		Header: make(http.Header),
	}

	// Parse the request
	var searchReq SearchRequest
	err := r.Parse(req, &searchReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Query: %s\n", searchReq.Query)
	fmt.Printf("Category: %s\n", searchReq.Category)
	fmt.Printf("Tags: %v\n", searchReq.Tags)
	fmt.Printf("Page: %d\n", searchReq.Page)
	fmt.Printf("Limit: %d\n", searchReq.Limit)

	// Output:
	// Query: golang
	// Category: programming
	// Tags: [web api http]
	// Page: 1
	// Limit: 10
}

// ExampleWithDisabledSplit demonstrates using query parser with disabled splitting.
func ExampleWithDisabledSplit() {
	// Structure with a field that shouldn't be split
	type TimeRequest struct {
		Timestamp string `query:"timestamp"` // Date format with commas
	}

	// Create parser with splitting disabled
	queryParser := parser.NewQuery(parser.WithDisabledSplit())

	// Create roamer instance
	r := roamer.NewRoamer(
		roamer.WithParsers(queryParser),
	)

	// Create request with comma-containing value
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			RawQuery: "timestamp=Mon, 02 Jan 2006 15:04:05 MST",
		},
		Header: make(http.Header),
	}

	var timeReq TimeRequest
	err := r.Parse(req, &timeReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Timestamp: %s\n", timeReq.Timestamp)

	// Output:
	// Timestamp: Mon, 02 Jan 2006 15:04:05 MST
}

// ExampleWithSplitSymbol demonstrates using a custom split symbol.
func ExampleWithSplitSymbol() {
	// Structure expecting semicolon-separated values
	type FilterRequest struct {
		Colors []string `query:"colors"`
	}

	// Create parser with custom split symbol
	queryParser := parser.NewQuery(parser.WithSplitSymbol(";"))

	// Create roamer instance
	r := roamer.NewRoamer(
		roamer.WithParsers(queryParser),
	)

	// Create request with semicolon-separated values
	// Properly construct URL using url.Parse and query parameter manipulation
	rawURL, err := url.Parse("http://example.com")
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return
	}

	q := rawURL.Query()
	q.Add("colors", "red;green;blue")
	rawURL.RawQuery = q.Encode()

	req := &http.Request{
		Method: "GET",
		URL:    rawURL,
		Header: make(http.Header),
	}

	var filterReq FilterRequest
	err = r.Parse(req, &filterReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Colors: %v\n", filterReq.Colors)

	// Output:
	// Colors: [red green blue]
}

// ExampleNewHeader demonstrates creating and using a header parser.
func ExampleNewHeader() {
	// Define a structure with header tags
	type RequestInfo struct {
		UserAgent     string `header:"User-Agent"`
		Accept        string `header:"Accept"`
		Authorization string `header:"Authorization"`
		ContentType   string `header:"Content-Type"`
	}

	// Create a header parser
	headerParser := parser.NewHeader()

	// Create roamer instance
	r := roamer.NewRoamer(
		roamer.WithParsers(headerParser),
	)

	// Create request with headers
	req := &http.Request{
		Method: "POST",
		Header: http.Header{
			"User-Agent":    {"MyApp/1.0"},
			"Accept":        {"application/json"},
			"Authorization": {"Bearer token123"},
			"Content-Type":  {"application/json"},
		},
	}

	var reqInfo RequestInfo
	err := r.Parse(req, &reqInfo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("User-Agent: %s\n", reqInfo.UserAgent)
	fmt.Printf("Accept: %s\n", reqInfo.Accept)
	fmt.Printf("Authorization: %s\n", reqInfo.Authorization)
	fmt.Printf("Content-Type: %s\n", reqInfo.ContentType)

	// Output:
	// User-Agent: MyApp/1.0
	// Accept: application/json
	// Authorization: Bearer token123
	// Content-Type: application/json
}

// ExampleNewCookie demonstrates creating and using a cookie parser.
func ExampleNewCookie() {
	// Define a structure with cookie tags
	type UserSession struct {
		SessionID string `cookie:"session_id"`
		UserID    string `cookie:"user_id"`
		Theme     string `cookie:"theme"`
	}

	// Create a cookie parser
	cookieParser := parser.NewCookie()

	// Create roamer instance
	r := roamer.NewRoamer(
		roamer.WithParsers(cookieParser),
	)

	// Create request with cookies
	req := &http.Request{
		Method: "GET",
		Header: http.Header{
			"Cookie": {"session_id=abc123; user_id=user456; theme=dark"},
		},
	}

	var session UserSession
	err := r.Parse(req, &session)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Session ID: %s\n", session.SessionID)
	fmt.Printf("User ID: %s\n", session.UserID)
	fmt.Printf("Theme: %s\n", session.Theme)

	// Output:
	// Session ID: abc123
	// User ID: user456
	// Theme: dark
}
