package parser

import (
	"net/http"
	"reflect"
)

const (
	// TagPath path tag.
	TagPath = "path"
)

// PathValueFunc returns path variable value with name from http request.
type PathValueFunc = func(r *http.Request, name string) (string, bool)

// Path is a path parser.
type Path struct {
	valueFromPath PathValueFunc
}

// NewPath returns new path parser.
func NewPath(valueFromPath PathValueFunc) *Path {
	if valueFromPath == nil {
		valueFromPath = func(_ *http.Request, _ string) (string, bool) { return "", false }
	}

	return &Path{valueFromPath: valueFromPath}
}

// Parse parses path value from request.
func (p *Path) Parse(r *http.Request, tag reflect.StructTag, _ Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagPath)
	if !ok {
		return "", false
	}

	return p.valueFromPath(r, tagValue)
}

// Tag returns working tag.
func (p *Path) Tag() string {
	return TagPath
}
