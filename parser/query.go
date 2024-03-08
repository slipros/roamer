package parser

import (
	"net/http"
	"reflect"
	"strings"
)

const (
	// TagQuery query tag.
	TagQuery = "query"
	// SplitSymbol array split symbol.
	SplitSymbol   = ","
	cacheKeyQuery = "query"
)

// QueryOptionsFunc query options changer.
type QueryOptionsFunc func(*Query)

// WithDisabledSplit disable array splitting.
func WithDisabledSplit() QueryOptionsFunc {
	return func(q *Query) {
		q.split = false
	}
}

// WithSplitSymbol set array split symbol.
func WithSplitSymbol(splitSymbol string) QueryOptionsFunc {
	return func(q *Query) {
		q.splitSymbol = splitSymbol
	}
}

// Query query parser.
type Query struct {
	split       bool
	splitSymbol string
}

// NewQuery returns new query parser.
func NewQuery(opts ...QueryOptionsFunc) *Query {
	q := Query{split: true, splitSymbol: SplitSymbol}

	for _, opt := range opts {
		opt(&q)
	}

	return &q
}

// Parse parses query from request.
//
// If query is not found in cache it will be parsed from request url and cached.
func (q *Query) Parse(r *http.Request, tag reflect.StructTag) (any, bool) {
	tagValue, ok := tag.Lookup(TagQuery)
	if !ok {
		return "", false
	}

	query := r.URL.Query()
	values, ok := query[tagValue]
	if !ok {
		return "", false
	}

	if len(values) == 1 {
		if q.split && strings.Contains(values[0], q.splitSymbol) {
			return strings.Split(values[0], q.splitSymbol), true
		}

		return values[0], true
	}

	return values, true
}

// Tag returns working tag.
func (q *Query) Tag() string {
	return TagQuery
}
