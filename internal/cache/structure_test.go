package cache

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStructureCache(t *testing.T) {
	decoders := []string{"json", "xml"}
	parsers := []string{"query", "header"}
	formatters := []string{"string"}

	sc := NewStructureCache(decoders, parsers, formatters)

	require.NotNil(t, sc)
	assert.Equal(t, decoders, sc.decoders)
	assert.Equal(t, parsers, sc.parsers)
	assert.Equal(t, formatters, sc.formatters)
}

func TestStructureCache_Fields_Successfully(t *testing.T) {
	type testStruct struct {
		Name       string `query:"name" header:"X-Name"`
		Age        int    `query:"age" default:"30"`
		Email      string `json:"email"`
		unexported string `query:"ignored"` // Should be ignored
		NoTags     string
	}

	type emptyStruct struct{}

	tests := []struct {
		name           string
		targetType     reflect.Type
		decoders       []string
		parsers        []string
		formatters     []string
		expectedFields []Field
	}{
		{
			name:       "basic struct with various tags",
			targetType: reflect.TypeOf(testStruct{}),
			decoders:   []string{"json"},
			parsers:    []string{"query", "header"},
			formatters: []string{"string"},
			expectedFields: []Field{
				{
					Index:       0,
					Name:        "Name",
					StructField: reflect.TypeOf(testStruct{}).Field(0),
					HasDefault:  false,
					Parsers:     []string{"query", "header"},
				},
				{
					Index:        1,
					Name:         "Age",
					StructField:  reflect.TypeOf(testStruct{}).Field(1),
					HasDefault:   true,
					DefaultValue: "30",
					Parsers:      []string{"query"},
				},
				{
					Index:       2,
					Name:        "Email",
					StructField: reflect.TypeOf(testStruct{}).Field(2),
					HasDefault:  false,
					Decoders:    []string{"json"},
				},
			},
		},
		{
			name:           "struct with no tags",
			targetType:     reflect.TypeOf(struct{ F1 string }{}),
			decoders:       []string{"json"},
			parsers:        []string{"query"},
			formatters:     []string{"string"},
			expectedFields: []Field{},
		},
		{
			name:           "empty struct",
			targetType:     reflect.TypeOf(emptyStruct{}),
			decoders:       []string{"json"},
			parsers:        []string{"query"},
			formatters:     []string{"string"},
			expectedFields: []Field{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewStructureCache(tt.decoders, tt.parsers, tt.formatters)

			// First call - should analyze and cache
			fields1 := sc.Fields(tt.targetType)

			// Custom comparison because reflect.StructField contains unexported fields
			require.Equal(t, len(tt.expectedFields), len(fields1))
			for i, expected := range tt.expectedFields {
				actual := fields1[i]
				assert.Equal(t, expected.Index, actual.Index)
				assert.Equal(t, expected.Name, actual.Name)
				assert.Equal(t, expected.HasDefault, actual.HasDefault)
				assert.Equal(t, expected.DefaultValue, actual.DefaultValue)
				assert.ElementsMatch(t, expected.Decoders, actual.Decoders)
				assert.ElementsMatch(t, expected.Parsers, actual.Parsers)
				assert.ElementsMatch(t, expected.Formatters, actual.Formatters)
			}

			// Second call - should return from cache
			fields2 := sc.Fields(tt.targetType)
			assert.Equal(t, fields1, fields2, "Second call should return identical data from cache")
		})
	}
}
