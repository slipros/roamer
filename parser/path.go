package parser

import (
	"net/http"
	"reflect"
)

const (
	// TagPath path tag.
	TagPath = "path"
)

// PathValueFunc returns path variable value.
type PathValueFunc = func(name string, r *http.Request) (string, bool)

type Path struct {
	valueFromPath PathValueFunc
}

// NewPath returns new path parser.
func NewPath(valueFromPath PathValueFunc) *Path {
	if valueFromPath == nil {
		valueFromPath = func(_ string, _ *http.Request) (string, bool) { return "", false }
	}

	return &Path{valueFromPath: valueFromPath}
}

// Tag returns working tag.
func (p *Path) Tag() string {
	return TagPath
}

// Parse parse path.
func (p *Path) Parse(r *http.Request, tag reflect.StructTag, _ Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagPath)
	if !ok {
		return nil, false
	}

	return p.valueFromPath(tagValue, r)
}
