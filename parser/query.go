package parser

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	// TagQuery query tag.
	TagQuery      = "query"
	cacheKeyQuery = "query"
)

// Query header parser.
type Query struct {
	splitSymbol string
}

// NewQuery returns new header parser.
func NewQuery(splitSymbol string) *Query {
	return &Query{splitSymbol: splitSymbol}
}

// Tag returns working tag.
func (q *Query) Tag() string {
	return TagQuery
}

// Parse parse query.
func (q *Query) Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagQuery)
	if !ok {
		return nil, false
	}

	query, ok := cache[cacheKeyQuery].(url.Values)
	if !ok {
		query = r.URL.Query()
		cache[cacheKeyQuery] = query
	}

	var split bool
	if before, found := strings.CutSuffix(tagValue, ",split"); found {
		tagValue = before
		split = true
	}

	values, ok := query[tagValue]
	if !ok {
		return nil, false
	}

	if len(values) == 1 {
		if split && strings.Contains(values[0], q.splitSymbol) {
			return strings.Split(values[0], q.splitSymbol), true
		}

		return values[0], true
	}

	return values, true
}
