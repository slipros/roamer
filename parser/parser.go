// Package parser provides components for extracting data from HTTP requests.
// These parsers extract values from different request sources based on struct tags.
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
