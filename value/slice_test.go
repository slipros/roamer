package value

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSliceString_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (reflect.Value, []string)
		expected interface{}
		options  []SliceOption
	}{
		{
			name: "string slice to string",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags string
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{"tag1", "tag2", "tag3"}
			},
			expected: "tag1,tag2,tag3",
		},
		{
			name: "string slice to string with custom separator",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags string
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{"tag1", "tag2", "tag3"}
			},
			expected: "tag1|tag2|tag3",
			options:  []SliceOption{WithSeparator("|")},
		},
		{
			name: "string slice to []string",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags []string
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{"tag1", "tag2", "tag3"}
			},
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "string slice to []any",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags []any
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{"tag1", "tag2", "tag3"}
			},
			expected: []any{"tag1", "tag2", "tag3"},
		},
		{
			name: "string slice to any (interface{})",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags any
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{"tag1", "tag2", "tag3"}
			},
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "numeric string slice to []int",
			setup: func() (reflect.Value, []string) {
				var s struct {
					IDs []int
				}
				return reflect.ValueOf(&s).Elem().FieldByName("IDs"), []string{"1", "2", "3"}
			},
			expected: []int{1, 2, 3},
		},
		{
			name: "numeric string slice to []float64",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Values []float64
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Values"), []string{"1.1", "2.2", "3.3"}
			},
			expected: []float64{1.1, 2.2, 3.3},
		},
		{
			name: "boolean string slice to []bool",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Flags []bool
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Flags"), []string{"true", "false", "true"}
			},
			expected: []bool{true, false, true},
		},
		{
			name: "string slice to *[]string (nil pointer)",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags *[]string
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{"tag1", "tag2", "tag3"}
			},
			expected: &[]string{"tag1", "tag2", "tag3"},
		},
		{
			name: "empty string slice to []string",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Tags []string
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Tags"), []string{}
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, values := tt.setup()
			err := SetSliceString(field, values, tt.options...)
			require.NoError(t, err)

			actualValue := field.Interface()

			// Special case for pointer types
			if field.Kind() == reflect.Pointer {
				if field.IsNil() {
					t.Fatalf("Expected pointer to be initialized, but it's nil")
				}

				// For pointers, we need to access the value it points to
				elem := field.Elem().Interface()

				// Compare the actual value with expected for pointer
				switch expected := tt.expected.(type) {
				case *[]string:
					actual, ok := elem.([]string)
					require.True(t, ok, "Expected []string but got different type")
					assert.Equal(t, *expected, actual)
				default:
					assert.Equal(t, tt.expected, field.Interface())
				}
				return
			}

			// Handle float slices with special comparison for approximate equality
			if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Float64 {
				expectedFloats, ok := tt.expected.([]float64)
				if ok {
					actualFloats, ok := actualValue.([]float64)
					require.True(t, ok, "Expected []float64 but got different type")
					require.Equal(t, len(expectedFloats), len(actualFloats))

					for i := range expectedFloats {
						assert.InDelta(t, expectedFloats[i], actualFloats[i], 1e-7,
							"Float values at index %d should be approximately equal", i)
					}
					return
				}
			}

			assert.Equal(t, tt.expected, actualValue, "Value should be correctly set")
		})
	}
}

func TestSetSliceString_Failure(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (reflect.Value, []string)
		errorMsg string
	}{
		{
			name: "string slice to unsupported type (map)",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Map map[string]string
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Map"), []string{"tag1", "tag2", "tag3"}
			},
			errorMsg: "cannot convert []string to field of type",
		},
		{
			name: "string slice to struct",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Inner struct {
						Field string
					}
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Inner"), []string{"tag1", "tag2", "tag3"}
			},
			errorMsg: "cannot convert []string to field of type",
		},
		{
			name: "invalid numeric conversion to []int",
			setup: func() (reflect.Value, []string) {
				var s struct {
					IDs []int
				}
				return reflect.ValueOf(&s).Elem().FieldByName("IDs"), []string{"1", "not-a-number", "3"}
			},
			errorMsg: "failed to convert string",
		},
		{
			name: "invalid boolean conversion to []bool",
			setup: func() (reflect.Value, []string) {
				var s struct {
					Flags []bool
				}
				return reflect.ValueOf(&s).Elem().FieldByName("Flags"), []string{"true", "not-a-bool", "false"}
			},
			errorMsg: "failed to convert string",
		},
		{
			name: "string slice to non-settable field",
			setup: func() (reflect.Value, []string) {
				type privateStruct struct {
					tags []string
				}
				s := &privateStruct{}
				return reflect.ValueOf(s).Elem().FieldByName("tags"), []string{"tag1", "tag2", "tag3"}
			},
			errorMsg: "is not settable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, values := tt.setup()
			err := SetSliceString(field, values)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// BenchmarkSetSliceString_StringToString benchmarks converting a string slice to a string
func BenchmarkSetSliceString_StringToString(b *testing.B) {
	var s struct {
		Tags string
	}
	field := reflect.ValueOf(&s).Elem().FieldByName("Tags")
	values := []string{"tag1", "tag2", "tag3", "tag4", "tag5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetSliceString(field, values)
	}
}

// BenchmarkSetSliceString_StringToSliceString benchmarks converting a string slice to a []string
func BenchmarkSetSliceString_StringToSliceString(b *testing.B) {
	var s struct {
		Tags []string
	}
	field := reflect.ValueOf(&s).Elem().FieldByName("Tags")
	values := []string{"tag1", "tag2", "tag3", "tag4", "tag5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetSliceString(field, values)
	}
}

// BenchmarkSetSliceString_StringToSliceInt benchmarks converting a string slice to an []int
func BenchmarkSetSliceString_StringToSliceInt(b *testing.B) {
	var s struct {
		IDs []int
	}
	field := reflect.ValueOf(&s).Elem().FieldByName("IDs")
	values := []string{"1", "2", "3", "4", "5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetSliceString(field, values)
	}
}

// BenchmarkSetSliceString_WithSeparator benchmarks using a custom separator
func BenchmarkSetSliceString_WithSeparator(b *testing.B) {
	var s struct {
		Tags string
	}
	field := reflect.ValueOf(&s).Elem().FieldByName("Tags")
	values := []string{"tag1", "tag2", "tag3", "tag4", "tag5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SetSliceString(field, values, WithSeparator("|"))
	}
}
