package roamer

import (
	"reflect"
)

// Formatter is a formatter.
//
//go:generate mockery --name=Formatter --outpkg=mock --output=./mock
type Formatter interface {
	Format(tag reflect.StructTag, ptr any) error
	Tag() string
}

// Formatters is a map of formatters where keys are tags for given formatters.
type Formatters map[string]Formatter

func (ft Formatters) has(tag reflect.StructTag) bool {
	for t := range ft {
		if _, ok := tag.Lookup(t); ok {
			return true
		}
	}

	return false
}
