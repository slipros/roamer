package parser

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	// TagQuery is the struct tag name used for parsing query parameters.
	// Fields tagged with this will be populated from matching URL query parameters.
	// Example: `query:"user_id"`
	TagQuery = "query"

	// SplitSymbol is the default character used to split comma-separated query values
	// when the split feature is enabled.
	SplitSymbol = ","

	// cacheKeyQuery is the key used to store parsed query parameters in the cache.
	cacheKeyQuery = "query"
)

// QueryOptionsFunc is a function type for configuring a Query parser instance.
// It follows the functional options pattern to provide a clean and extensible API.
type QueryOptionsFunc func(*Query)

// WithDisabledSplit disables the automatic splitting of comma-separated query values.
// By default, if a query parameter contains comma-separated values and splitting is enabled,
// the parser will split it into a slice. This option disables that behavior.
//
// Example:
//
//	// Create a query parser with splitting disabled
//	parser := parser.NewQuery(parser.WithDisabledSplit())
//
//	// With splitting disabled, a query like "?tags=foo,bar,baz" will be parsed
//	// as a single string "foo,bar,baz" rather than a slice ["foo", "bar", "baz"]
func WithDisabledSplit() QueryOptionsFunc {
	return func(q *Query) {
		q.split = false
	}
}

// WithSplitSymbol sets the character used to split query values.
// By default, the parser uses a comma (,) as the split symbol.
// This option allows using a different character instead.
//
// Example:
//
//	// Create a query parser that splits on semicolons instead of commas
//	parser := parser.NewQuery(parser.WithSplitSymbol(";"))
//
//	// With this configuration, a query like "?tags=foo;bar;baz" will be parsed
//	// as a slice ["foo", "bar", "baz"]
func WithSplitSymbol(splitSymbol string) QueryOptionsFunc {
	return func(q *Query) {
		q.splitSymbol = splitSymbol
	}
}

// Query is a parser for extracting URL query parameters from HTTP requests.
// It can handle single values, multiple values (using repeated parameters),
// and comma-separated values.
type Query struct {
	split       bool   // Whether to split comma-separated values
	splitSymbol string // The character to use when splitting values
}

// NewQuery creates a Query parser with specified options.
// By default, splits comma-separated values into slices.
//
// Example:
//
//	// Default query parser
//	parser := parser.NewQuery()
//
//	// With custom options
//	parser := parser.NewQuery(
//	    parser.WithDisabledSplit(),      // Don't split comma-separated values
//	    parser.WithSplitSymbol(";"),     // Use semicolon as separator
//	)
func NewQuery(opts ...QueryOptionsFunc) *Query {
	q := Query{split: true, splitSymbol: SplitSymbol}

	for _, opt := range opts {
		opt(&q)
	}

	return &q
}

// Parse extracts a query parameter from an HTTP request based on the provided struct tag.
// If the query parameter exists, it returns the value and true.
// If the query parameter does not exist, it returns an empty string and false.
//
// For efficiency, the parser caches the parsed query parameters to avoid
// parsing them multiple times for different struct fields.
//
// Parameters:
//   - r: The HTTP request to extract query parameters from.
//   - tag: The struct tag containing the query parameter name.
//   - cache: A cache for storing parsed query parameters.
//
// Returns:
//   - any: The parsed query parameter (string, []string, or split string).
//   - bool: Whether the query parameter was found.
func (q *Query) Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagQuery)
	if !ok {
		return "", false
	}

	query, ok := cache[cacheKeyQuery].(url.Values)
	if !ok {
		query = r.URL.Query()
		cache[cacheKeyQuery] = query
	}

	values, ok := query[tagValue]
	if !ok {
		return "", false
	}

	if len(values) == 1 {
		v := values[0]

		if q.split && strings.Contains(v, q.splitSymbol) {
			return strings.Split(v, q.splitSymbol), true
		}

		return v, true
	}

	return values, true
}

// Tag returns the name of the struct tag that this parser handles.
// For the Query parser, this is "query".
func (q *Query) Tag() string {
	return TagQuery
}
