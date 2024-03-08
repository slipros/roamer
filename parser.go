package roamer

import (
	"net/http"
	"reflect"
)

// Parser is a parser.
//
//go:generate mockery --name=Parser --outpkg=mock --output=./mock
type Parser interface {
	Parse(r *http.Request, tag reflect.StructTag) (any, bool)
	Tag() string
}

// Parsers is a map of parsers where keys are tags for given parsers.
type Parsers map[string]Parser
