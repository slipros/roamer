package roamer

import (
	"net/http"
	"reflect"
)

// Parser parser.
//
//go:generate mockery --name=Parser --outpkg=mock --output=./mock
type Parser interface {
	Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool)
	Tag() string
}

// Parsers is a map of parsers where keys are tags for given parsers.
type Parsers map[string]Parser

// Cache is a cache of parsed values.
type Cache = map[string]any
