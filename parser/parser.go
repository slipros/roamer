// Package parser provides parsers for extracting data from HTTP requests.
// These parsers are used by the roamer package to populate struct fields
// based on struct tags.
package parser

// Cache is a type alias for a map that stores parsed values.
// It is used to avoid redundant parsing of the same values from an HTTP request
// when multiple struct fields use the same tag.
//
// For example, if multiple struct fields use the same query parameter,
// the parameter will be parsed once and cached for subsequent fields.
//
// Example usage (internal to parsers):
//
//	func (p *MyParser) Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool) {
//	    // Check if value is already cached
//	    if cachedValue, ok := cache["my_cache_key"]; ok {
//	        return cachedValue, true
//	    }
//
//	    // Parse value
//	    // ...
//
//	    // Cache value for future use
//	    cache["my_cache_key"] = parsedValue
//
//	    return parsedValue, true
//	}
type Cache = map[string]any
