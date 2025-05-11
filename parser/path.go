// Package parser provides parsers for extracting data from HTTP requests.
package parser

import (
	"net/http"
	"reflect"
)

const (
	// TagPath is the struct tag name used for parsing URL path parameters.
	// Fields tagged with this will be populated from matching path parameters.
	// Example: `path:"user_id"`
	TagPath = "path"
)

// PathValueFunc is a function type that extracts path parameters from HTTP requests.
// It takes an HTTP request and a parameter name, and returns the parameter value
// and a boolean indicating whether the parameter was found.
//
// Different routers have different ways of storing path parameters, so this
// function type allows the Path parser to work with any router by providing
// the appropriate adapter function.
type PathValueFunc = func(r *http.Request, name string) (string, bool)

// Path is a parser for extracting URL path parameters from HTTP requests.
// It delegates the actual extraction to a provided function, making it
// compatible with any HTTP router.
type Path struct {
	valueFromPath PathValueFunc // Function to extract path parameters
}

// NewPath creates a new Path parser with the specified extraction function.
// The extraction function should be provided by the router being used.
//
// If no function is provided (nil), a default function that always returns
// empty values is used.
//
// Example:
//
//	// Using with standard ServeMux (Go 1.22+)
//	pathParser := parser.NewPath(parser.ServeMuxValueFromPath)
//
//	// Using with chi router
//	import "github.com/go-chi/chi/v5"
//	// ...
//	router := chi.NewRouter()
//	pathParser := parser.NewPath(func(r *http.Request, name string) (string, bool) {
//	    value := chi.URLParam(r, name)
//	    if value == "" {
//	        return "", false
//	    }
//	    return value, true
//	})
//
//	// Using with gorilla/mux
//	import "github.com/gorilla/mux"
//	// ...
//	pathParser := parser.NewPath(func(r *http.Request, name string) (string, bool) {
//	    vars := mux.Vars(r)
//	    value, ok := vars[name]
//	    return value, ok
//	})
func NewPath(valueFromPath PathValueFunc) *Path {
	if valueFromPath == nil {
		valueFromPath = func(_ *http.Request, _ string) (string, bool) { return "", false }
	}

	return &Path{valueFromPath: valueFromPath}
}

// Parse extracts a path parameter from an HTTP request based on the provided struct tag.
// It delegates the actual extraction to the valueFromPath function provided
// during initialization.
//
// Parameters:
//   - r: The HTTP request to extract path parameters from.
//   - tag: The struct tag containing the path parameter name.
//   - _: Cache parameter (not used by this parser).
//
// Returns:
//   - any: The parsed path parameter (string).
//   - bool: Whether the path parameter was found.
func (p *Path) Parse(r *http.Request, tag reflect.StructTag, _ Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagPath)
	if !ok {
		return "", false
	}

	return p.valueFromPath(r, tagValue)
}

// Tag returns the name of the struct tag that this parser handles.
// For the Path parser, this is "path".
func (p *Path) Tag() string {
	return TagPath
}

// ServeMuxValueFromPath is a PathValueFunc implementation for the standard
// Go 1.22+ ServeMux with native path parameter support.
//
// Example:
//
//	// Create a path parser for standard ServeMux
//	pathParser := parser.NewPath(parser.ServeMuxValueFromPath)
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(pathParser),
//	)
//
//	// In your HTTP handler
//	http.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	    var req struct {
//	        UserID string `path:"id"`
//	    }
//	    if err := r.Parse(r, &req); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Use req.UserID...
//	})
func ServeMuxValueFromPath(r *http.Request, name string) (string, bool) {
	value := r.PathValue(name)
	if len(value) == 0 {
		return "", false
	}

	return value, true
}
