package roamer

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
)

// KitchenSink is a struct that uses a wide variety of tags and types
// to test the full functionality of the roamer library in an end-to-end scenario.

type KitchenSink struct {
	// Parsers
	Header      string    `header:"X-Test-Header"`
	Query       int       `query:"id"`
	QuerySlice  []string  `query:"tags"`
	Cookie      string    `cookie:"session_id"`
	Path        string    `path:"userId"`
	Default     string    `default:"default_value"`
	QueryTime   time.Time `query:"timestamp"`
	HeaderFloat float64   `header:"X-Float-Value"`

	// Decoder fields (for JSON, XML, Form)
	Name     string   `json:"name" xml:"name" form:"name"`
	Age      int      `json:"age" xml:"age" form:"age"`
	IsActive bool     `json:"isActive" xml:"isActive" form:"isActive"`
	Friends  []string `json:"friends" xml:"friends" form:"friends"`
}

// FuzzRoamerEndToEnd performs an end-to-end fuzz test of the Roamer.Parse function.
// It constructs a complex Roamer instance with all available parsers and decoders
// and attempts to parse a request into a comprehensive struct (KitchenSink).
// The fuzzer generates data for all parts of the HTTP request.
func FuzzRoamerEndToEnd(f *testing.F) {
	// Add seed values to guide the fuzzer.
	f.Add(
		"id=123&tags=go,fuzzing&timestamp=2025-08-30T12:00:00Z", // query
		"X-Test-Header:hello_world",                             // header
		"session_id=abc-123",                                    // cookie
		"/users/user-456/profile",                               // path
		`{"name":"John Doe","age":30,"isActive":true,"friends":["Jane","Tom"]}`,
		"application/json",
	)
	f.Add(
		"id=999",
		"X-Float-Value:3.14",
		"",
		"/items/item-789",
		`<KitchenSink><name>Jane XML</name><age>25</age><isActive>false</isActive></KitchenSink>`,
		"application/xml",
	)
	f.Add(
		"", "", "", "/",
		"name=Form+User&age=40&isActive=true&friends=bob&friends=alice",
		"application/x-www-form-urlencoded",
	)
	f.Add("invalid-query", "Invalid:Header", "invalid=cookie", "/path", "not-a-body", "text/plain")
	// Create a fully configured Roamer instance once.
	// This instance includes all standard parsers, decoders, and formatters.
	pathParser := parser.NewPath(func(r *http.Request, key string) (string, bool) {
		// Simple mock path extractor for testing.
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 {
			if key == "userId" {
				return parts[2], true
			}
		}
		return "", false
	})

	roamerInstance := NewRoamer(
		WithParsers(
			parser.NewHeader(),
			parser.NewQuery(),
			parser.NewCookie(),
			pathParser,
		),
		WithDecoders(
			decoder.NewJSON(),
			decoder.NewXML(),
			decoder.NewFormURL(),
		),
		WithFormatters(
			formatter.NewTime(),
			formatter.NewSlice(),
			formatter.NewNumeric(),
			formatter.NewString(),
		),
	)

	f.Fuzz(func(t *testing.T, query, header, cookie, path, body, contentType string) {
		// Construct the HTTP request from fuzzed data.
		reqURL := "http://example.com" + path + "?" + query
		req, err := http.NewRequest("POST", reqURL, bytes.NewReader([]byte(body)))
		if err != nil {
			return // Skip if fuzzer generates invalid request data
		}

		// Set header
		if parts := strings.SplitN(header, ":", 2); len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
		req.Header.Set("Content-Type", contentType)

		// Set cookie
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		// The goal is to find a combination of inputs that causes a panic.
		// We don't need to validate the output, just ensure it doesn't crash.
		var dest KitchenSink
		_ = roamerInstance.Parse(req, &dest)
	})
}
