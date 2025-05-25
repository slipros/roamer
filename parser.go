package roamer

import (
	"net/http"
	"reflect"

	"github.com/slipros/roamer/parser"
)

// Parser is an interface for components that extract data from specific parts
// of an HTTP request based on struct tags. Different parsers can handle different
// parts of the request (headers, query parameters, cookies, path parameters, etc.).
//
// Implementing a custom parser allows extending the functionality of the roamer
// package to support additional data sources or custom parsing logic.
//
//go:generate mockery --name=Parser --outpkg=mockroamer --output=./mockroamer
type Parser interface {
	// Parse extracts data from an HTTP request based on the provided struct tag.
	// It returns the parsed value and a boolean indicating whether parsing was successful.
	// The cache parameter can be used to store intermediate results for performance optimization.
	//
	// Parameters:
	//   - r: The HTTP request to parse data from.
	//   - tag: The struct tag containing parsing instructions.
	//   - cache: A cache for storing intermediate parsing results.
	//
	// Returns:
	//   - any: The parsed value (can be of any type).
	//   - bool: Whether parsing was successful (false if no matching data was found).
	Parse(r *http.Request, tag reflect.StructTag, cache parser.Cache) (any, bool)

	// Tag returns the name of the struct tag that this parser handles.
	// For example, a query parameter parser might return "query",
	// a header parser might return "header", etc.
	Tag() string
}

// Parsers is a map of registered parsers where keys are the tag names
// returned by the Parser.Tag() method.
type Parsers map[string]Parser
