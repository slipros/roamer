package roamer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
)

// BenchmarkStruct represents a comprehensive benchmark structure for testing
// various parsing scenarios and performance characteristics
type BenchmarkStruct struct {
	// Query parameters
	QueryString      string    `query:"qstring"`
	QueryInt         int       `query:"qint"`
	QueryFloat       float64   `query:"qfloat"`
	QueryBool        bool      `query:"qbool"`
	QueryTime        time.Time `query:"qtime"`
	QueryStringSlice []string  `query:"qslice"`
	QueryIntSlice    []int     `query:"qintslice"`

	// HTTP Headers
	HeaderAuth      string `header:"Authorization"`
	HeaderUserAgent string `header:"User-Agent"`
	HeaderCustom    string `header:"X-Custom-Header"`
	HeaderInt       int    `header:"X-Int-Header"`

	// Cookies
	CookieSession string `cookie:"session_id"`
	CookiePrefs   string `cookie:"user_prefs"`
	CookieInt     int    `cookie:"numeric_cookie"`

	// JSON body fields
	JSONName     string            `json:"name"`
	JSONEmail    string            `json:"email"`
	JSONAge      int               `json:"age"`
	JSONActive   bool              `json:"active"`
	JSONTags     []string          `json:"tags"`
	JSONMetadata map[string]string `json:"metadata"`

	// Formatted fields with string formatter
	FormattedLower string `json:"formatted_lower" format:"lower_case"`
	FormattedTrim  string `json:"formatted_trim" format:"trim_space"`
	FormattedMulti string `json:"formatted_multi" format:"trim_space,lower_case"`
}

// Comprehensive benchmark testing complete parse pipeline with all components
func BenchmarkParse_CompletePipeline(b *testing.B) {
	req := createComprehensiveTestRequest(b)

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(
			parser.NewQuery(),
			parser.NewHeader(),
			parser.NewCookie(),
		),
		WithFormatters(formatter.NewString()),
	)

	var target BenchmarkStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := roamer.Parse(req, &target); err != nil {
			b.Fatal(err)
		}
		// Reset struct to simulate fresh parsing
		target = BenchmarkStruct{}
	}
}

// Benchmark Parse with pre-filled struct and SkipFilled option
func BenchmarkParse_SkipFilledAdvanced(b *testing.B) {
	req := createComprehensiveTestRequest(b)

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(
			parser.NewQuery(),
			parser.NewHeader(),
			parser.NewCookie(),
		),
		WithSkipFilled(true),
	)

	// Pre-fill half the fields to test skip behavior
	target := BenchmarkStruct{
		QueryString:   "prefilled",
		JSONName:      "prefilled",
		HeaderAuth:    "Bearer prefilled",
		CookieSession: "prefilled_session",
	}
	originalTarget := target

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := roamer.Parse(req, &target); err != nil {
			b.Fatal(err)
		}
		// Reset to pre-filled state
		target = originalTarget
	}
}

// Benchmark varying struct complexity
func BenchmarkParse_StructComplexity(b *testing.B) {
	tests := []struct {
		name   string
		target any
		setup  func() *http.Request
	}{
		{
			name:   "SmallStruct_5Fields",
			target: &SmallStruct{},
			setup:  func() *http.Request { return createSimpleTestRequest(b, 5) },
		},
		{
			name:   "MediumStruct_15Fields",
			target: &MediumStruct{},
			setup:  func() *http.Request { return createMediumTestRequest(b) },
		},
		{
			name:   "LargeStruct_30Fields",
			target: &LargeStruct{},
			setup:  func() *http.Request { return createLargeTestRequest(b) },
		},
		{
			name:   "BenchmarkStruct_25Fields",
			target: &BenchmarkStruct{},
			setup:  func() *http.Request { return createComprehensiveTestRequest(b) },
		},
	}

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
	)

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := tt.setup()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if err := roamer.Parse(req, tt.target); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark memory allocation patterns with different request sizes
func BenchmarkParse_MemoryAllocation(b *testing.B) {
	tests := []struct {
		name        string
		queryParams int
		headers     int
		jsonFields  int
	}{
		{"Small_5_5_5", 5, 5, 5},
		{"Medium_20_10_15", 20, 10, 15},
		{"Large_100_25_50", 100, 25, 50},
		{"XLarge_500_50_100", 500, 50, 100},
	}

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
	)

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := createScalableTestRequest(b, tt.queryParams, tt.headers, tt.jsonFields)
			var target map[string]any

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				target = make(map[string]any)
				if err := roamer.Parse(req, &target); err != nil {
					// Map parsing might fail, that's ok for memory allocation testing
				}
			}
		})
	}
}

// Benchmark concurrent parsing to test thread safety and performance
func BenchmarkParse_Concurrent(b *testing.B) {
	req := createComprehensiveTestRequest(b)

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader(), parser.NewCookie()),
		WithFormatters(formatter.NewString()),
	)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var target BenchmarkStruct
		for pb.Next() {
			if err := roamer.Parse(req, &target); err != nil {
				b.Fatal(err)
			}
			target = BenchmarkStruct{}
		}
	})
}

// Benchmark cache efficiency - repeated parsing of same struct type
func BenchmarkParse_CacheEfficiency(b *testing.B) {
	req := createComprehensiveTestRequest(b)

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
	)

	b.ReportAllocs()
	b.ResetTimer()

	// First parse should populate cache
	var target BenchmarkStruct
	if err := roamer.Parse(req, &target); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Subsequent parses should benefit from cache
	for i := 0; i < b.N; i++ {
		target = BenchmarkStruct{}
		if err := roamer.Parse(req, &target); err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions for creating test requests

func createComprehensiveTestRequest(b *testing.B) *http.Request {
	b.Helper()

	// Create JSON body
	jsonData := map[string]any{
		"name":   "John Doe",
		"email":  "john@example.com",
		"age":    30,
		"active": true,
		"tags":   []string{"developer", "golang", "api"},
		"metadata": map[string]string{
			"department": "engineering",
			"team":       "backend",
		},
		"formatted_lower": "  UPPERCASE TEXT  ",
		"formatted_trim":  "  trim me  ",
		"formatted_multi": "  TRIM AND LOWER  ",
	}

	jsonBytes, _ := json.Marshal(jsonData)

	// Create URL with query parameters
	u, _ := url.Parse("https://api.example.com/users")
	q := u.Query()
	q.Add("qstring", "query_value")
	q.Add("qint", "42")
	q.Add("qfloat", "3.14159")
	q.Add("qbool", "true")
	q.Add("qtime", time.Now().Format(time.RFC3339))
	q.Add("qslice", "value1,value2,value3")
	q.Add("qintslice", "1,2,3,4,5")
	u.RawQuery = q.Encode()

	// Create request
	req, _ := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBytes))

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("User-Agent", "BenchmarkClient/1.0")
	req.Header.Set("X-Custom-Header", "custom_value")
	req.Header.Set("X-Int-Header", "99")

	// Add cookies
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess_abc123"})
	req.AddCookie(&http.Cookie{Name: "user_prefs", Value: "theme=dark"})
	req.AddCookie(&http.Cookie{Name: "numeric_cookie", Value: "789"})

	return req
}

func createSimpleTestRequest(b *testing.B, fieldCount int) *http.Request {
	b.Helper()

	jsonData := map[string]any{
		"string": "test_value",
		"int":    123,
	}

	jsonBytes, _ := json.Marshal(jsonData)

	u, _ := url.Parse("https://api.example.com/simple")
	q := u.Query()
	for i := 0; i < fieldCount; i++ {
		q.Add(fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i))
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func createMediumTestRequest(b *testing.B) *http.Request {
	b.Helper()

	jsonData := map[string]any{
		"strings": []string{"str1", "str2", "str3"},
	}

	jsonBytes, _ := json.Marshal(jsonData)

	u, _ := url.Parse("https://api.example.com/medium")
	q := u.Query()
	q.Add("int", "123")
	q.Add("int8", "8")
	q.Add("int32", "32")
	q.Add("int64", "64")
	q.Add("time", time.Now().Format(time.RFC3339))
	q.Add("url", "https://example.com")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TestAgent/1.0")

	return req
}

func createLargeTestRequest(b *testing.B) *http.Request {
	b.Helper()

	jsonData := map[string]any{
		"string":       "large_test_value",
		"int":          999,
		"stringslice":  []string{"a", "b", "c", "d", "e"},
		"intslice":     []int{1, 2, 3, 4, 5},
		"float32slice": []float32{1.1, 2.2, 3.3},
		"boolslice":    []bool{true, false, true},
	}

	jsonBytes, _ := json.Marshal(jsonData)

	u, _ := url.Parse("https://api.example.com/large")
	q := u.Query()
	// Add many query parameters
	q.Add("string", "query_string")
	q.Add("int", "123")
	q.Add("int8", "8")
	q.Add("int16", "16")
	q.Add("int32", "32")
	q.Add("int64", "64")
	q.Add("uint", "123")
	q.Add("uint8", "8")
	q.Add("uint16", "16")
	q.Add("uint32", "32")
	q.Add("uint64", "64")
	q.Add("float32", "3.14")
	q.Add("float64", "3.141592653589793")
	q.Add("bool", "true")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-String", "header_value")
	req.Header.Set("X-Int", "456")

	return req
}

func createScalableTestRequest(b *testing.B, queryParams, headers, jsonFields int) *http.Request {
	b.Helper()

	// Create JSON with specified number of fields
	jsonData := make(map[string]any)
	for i := 0; i < jsonFields; i++ {
		jsonData[fmt.Sprintf("json_field_%d", i)] = fmt.Sprintf("json_value_%d", i)
	}

	jsonBytes, _ := json.Marshal(jsonData)

	// Create URL with specified number of query parameters
	u, _ := url.Parse("https://api.example.com/scalable")
	q := u.Query()
	for i := 0; i < queryParams; i++ {
		q.Add(fmt.Sprintf("query_param_%d", i), fmt.Sprintf("query_value_%d", i))
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add specified number of headers
	for i := 0; i < headers; i++ {
		req.Header.Set(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("header_value_%d", i))
	}

	return req
}
