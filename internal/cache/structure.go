package cache

import (
	"reflect"
	"sync"
)

// Field represents a cached struct field with metadata about its properties
// and applicable parsers, decoders, and formatters.
type Field struct {
	Index        int
	Name         string
	StructField  reflect.StructField
	HasDefault   bool
	DefaultValue string
	Decoders     []string
	Formatters   []string
	Parsers      []string
}

// StructureCache provides thread-safe caching of struct field analysis
// to avoid repeated reflection operations on the same types.
type StructureCache struct {
	cache                         sync.Map
	decoders, parsers, formatters []string
}

// NewStructureCache creates a new structure cache with the specified
// decoders, parsers, and formatters for field analysis.
func NewStructureCache(decoders, parsers, formatters []string) StructureCache {
	return StructureCache{
		decoders:   decoders,
		parsers:    parsers,
		formatters: formatters,
	}
}

// Fields returns the cached field analysis for the given type.
// If not cached, it performs the analysis and caches the result.
func (s *StructureCache) Fields(t reflect.Type) []Field {
	if value, ok := s.cache.Load(t); ok {
		return value.([]Field)
	}

	fields := s.analyzeStruct(t)

	// Try to store, but use existing value if another goroutine stored first
	if actual, loaded := s.cache.LoadOrStore(t, fields); loaded {
		return actual.([]Field)
	}

	return fields
}

// analyzeStruct performs reflection-based analysis of a struct type
// to extract field metadata and applicable tags.
func (s *StructureCache) analyzeStruct(t reflect.Type) []Field {
	fields := make([]Field, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if !f.IsExported() || len(f.Tag) == 0 {
			continue
		}

		defaultValue, hasDefault := f.Tag.Lookup("default")

		fields = append(fields, Field{
			Index:        i,
			Name:         f.Name,
			StructField:  f,
			HasDefault:   hasDefault,
			DefaultValue: defaultValue,
			Decoders:     s.tagLookup(f.Tag, s.decoders),
			Parsers:      s.tagLookup(f.Tag, s.parsers),
			Formatters:   s.tagLookup(f.Tag, s.formatters),
		})
	}

	return fields
}

// tagLookup checks which of the given tag names are present in the struct tag
// and returns a slice of the found tag names.
func (s *StructureCache) tagLookup(tag reflect.StructTag, values []string) []string {
	exists := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := tag.Lookup(v); !ok {
			continue
		}

		exists = append(exists, v)
	}

	return exists
}
