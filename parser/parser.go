// Package parser provides components for extracting data from different parts of HTTP requests.
//
// Parsers implement the Parser interface and are responsible for extracting values from
// specific request sources (query parameters, headers, cookies, path variables) based on
// struct tags. Each parser handles a specific tag and request source.
//
// # Built-in Parsers
//
//   - Query: Extracts URL query parameters using the "query" tag
//   - Header: Extracts HTTP headers using the "header" tag
//   - Cookie: Extracts HTTP cookies using the "cookie" tag
//   - Path: Extracts URL path variables using the "path" tag (requires router integration)
//
// # Basic Usage
//
//	// Create parsers
//	queryParser := parser.NewQuery()
//	headerParser := parser.NewHeader()
//	cookieParser := parser.NewCookie()
//
//	// Use with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(queryParser, headerParser, cookieParser),
//	)
//
// # Path Parser Integration
//
// The Path parser requires a router-specific adapter function:
//
//	// Standard library ServeMux (Go 1.22+)
//	pathParser := parser.NewPath(parser.ServeMuxValueFromPath)
//
//	// Chi router
//	import chiParser "github.com/slipros/roamer/pkg/chi"
//	pathParser := parser.NewPath(chiParser.NewPath(router))
//
//	// Gorilla Mux
//	import gorillaParser "github.com/slipros/roamer/pkg/gorilla"
//	pathParser := parser.NewPath(gorillaParser.Path)
//
// # Custom Parsers
//
// Implement the Parser interface to support custom request sources:
//
//	type MyParser struct{}
//
//	func (p *MyParser) Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool) {
//	    // Custom parsing logic
//	    value := extractCustomValue(r)
//	    return value, true
//	}
//
//	func (p *MyParser) Tag() string {
//	    return "custom"
//	}
//
// # Thread Safety
//
// All built-in parsers are safe for concurrent use and should be reused
// across multiple requests for optimal performance.
//
// # Performance
//
// Parsers use a cache parameter to avoid redundant parsing of the same request
// elements. For example, Query parser caches parsed query parameters to avoid
// re-parsing for each struct field.
package parser

// Cache stores parsed values to prevent redundant parsing of the same request elements.
// Used internally by parsers to optimize performance for repeated request data.
//
// Example usage within parsers:
//
//	func (p *MyParser) Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool) {
//	    if cachedValue, ok := cache["my_key"]; ok {
//	        return cachedValue, true
//	    }
//
//	    // Parse value and store in cache
//	    value := parseValue(r)
//	    cache["my_key"] = value
//
//	    return value, true
//	}
type Cache = map[string]any
