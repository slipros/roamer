// Package parser parse tags and return it's values.
package parser

import (
	"net/http"
	"reflect"
)

// Parser parser.
//
//go:generate mockery --name=Parser --outpkg=mockparser --output=./mockparser
type Parser interface {
	Tag() string
	Parse(r *http.Request, tag reflect.StructTag, cache Cache) (any, bool)
}

// Parsers parsers.
type Parsers map[string]Parser

// Cache cache.
type Cache = map[string]any
