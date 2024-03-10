package roamer

import (
	"net/http"
	"reflect"

	"github.com/slipros/roamer/parser"
)

// Parser is a parser.
//
//go:generate mockery --name=Parser --outpkg=mock --output=./mock
type Parser interface {
	Parse(r *http.Request, tag reflect.StructTag, cache parser.Cache) (any, bool)
	Tag() string
}

// Parsers is a map of parsers where keys are tags for given parsers.
type Parsers map[string]Parser
