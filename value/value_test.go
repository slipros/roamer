package value

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/stretchr/testify/require"
)

// Custom stringer type for testing
type CustomStringer struct {
	Value string
}

func (cs CustomStringer) String() string {
	return cs.Value
}

// Tests for successful value setting
func TestSet_Successfully(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
	}{
		// String tests
		{
			name: "set string to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     "test string",
			expected:  "test string",
		},
		{
			name: "set string to *string field",
			targetStructFn: func() any {
				return &struct{ Value *string }{}
			},
			fieldName: "Value",
			value:     "test string pointer",
			expected:  "test string pointer",
		},
		{
			name: "set *string to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value: func() *string {
				s := "string pointer value"
				return &s
			}(),
			expected: "string pointer value",
		},
		{
			name: "set nil *string to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     (*string)(nil),
			expected:  "",
		},

		// Boolean tests
		{
			name: "set bool to bool field",
			targetStructFn: func() any {
				return &struct{ Value bool }{}
			},
			fieldName: "Value",
			value:     true,
			expected:  true,
		},
		{
			name: "set bool to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     true,
			expected:  "true",
		},
		{
			name: "set *bool to bool field",
			targetStructFn: func() any {
				return &struct{ Value bool }{}
			},
			fieldName: "Value",
			value: func() *bool {
				b := true
				return &b
			}(),
			expected: true,
		},
		{
			name: "set nil *bool to bool field",
			targetStructFn: func() any {
				return &struct{ Value bool }{}
			},
			fieldName: "Value",
			value:     (*bool)(nil),
			expected:  false,
		},

		// Integer tests
		{
			name: "set int to int field",
			targetStructFn: func() any {
				return &struct{ Value int }{}
			},
			fieldName: "Value",
			value:     42,
			expected:  42,
		},
		{
			name: "set int to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     42,
			expected:  "42",
		},
		{
			name: "set int8 to int16 field",
			targetStructFn: func() any {
				return &struct{ Value int16 }{}
			},
			fieldName: "Value",
			value:     int8(42),
			expected:  int16(42),
		},
		{
			name: "set *int to int field",
			targetStructFn: func() any {
				return &struct{ Value int }{}
			},
			fieldName: "Value",
			value: func() *int {
				i := 42
				return &i
			}(),
			expected: 42,
		},
		{
			name: "set nil *int to int field",
			targetStructFn: func() any {
				return &struct{ Value int }{}
			},
			fieldName: "Value",
			value:     (*int)(nil),
			expected:  0,
		},

		// Unsigned integer tests
		{
			name: "set uint to uint field",
			targetStructFn: func() any {
				return &struct{ Value uint }{}
			},
			fieldName: "Value",
			value:     uint(42),
			expected:  uint(42),
		},
		{
			name: "set uint to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     uint(42),
			expected:  "42",
		},
		{
			name: "set *uint to uint field",
			targetStructFn: func() any {
				return &struct{ Value uint }{}
			},
			fieldName: "Value",
			value: func() *uint {
				u := uint(42)
				return &u
			}(),
			expected: uint(42),
		},
		{
			name: "set nil *uint to uint field",
			targetStructFn: func() any {
				return &struct{ Value uint }{}
			},
			fieldName: "Value",
			value:     (*uint)(nil),
			expected:  uint(0),
		},

		// Float tests
		{
			name: "set float64 to float64 field",
			targetStructFn: func() any {
				return &struct{ Value float64 }{}
			},
			fieldName: "Value",
			value:     3.14159,
			expected:  3.14159,
		},
		{
			name: "set float32 to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     float32(3.14),
			expected:  "3.140000104904175", // Exact representation of float32(3.14)
		},
		{
			name: "set *float64 to float64 field",
			targetStructFn: func() any {
				return &struct{ Value float64 }{}
			},
			fieldName: "Value",
			value: func() *float64 {
				f := 3.14159
				return &f
			}(),
			expected: 3.14159,
		},
		{
			name: "set nil *float64 to float64 field",
			targetStructFn: func() any {
				return &struct{ Value float64 }{}
			},
			fieldName: "Value",
			value:     (*float64)(nil),
			expected:  float64(0),
		},

		// Slice tests
		{
			name: "set []string to []string field",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName: "Value",
			value:     []string{"one", "two", "three"},
			expected:  []string{"one", "two", "three"},
		},
		{
			name: "set []string to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     []string{"one", "two", "three"},
			expected:  "one,two,three",
		},
		{
			name: "set []any to []string field",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName: "Value",
			value:     []any{"one", "two", "three"},
			expected:  []string{"one", "two", "three"},
		},
		{
			name: "set []any to []int field",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName: "Value",
			value:     []any{1, 2, 3},
			expected:  []int{1, 2, 3},
		},

		// Custom stringer tests
		{
			name: "set fmt.Stringer to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     CustomStringer{Value: "stringer value"},
			expected:  "stringer value",
		},

		// Nil value test
		{
			name: "set nil to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     nil,
			expected:  "",
		},
		{
			name: "set nil to *string field",
			targetStructFn: func() any {
				return &struct{ Value *string }{}
			},
			fieldName: "Value",
			value:     nil,
			expected:  (*string)(nil),
		},

		// Map tests
		{
			name: "set map to identical map field",
			targetStructFn: func() any {
				return &struct{ Value map[string]string }{}
			},
			fieldName: "Value",
			value:     map[string]string{"key": "value"},
			expected:  map[string]string{"key": "value"},
		},

		// Interface tests
		{
			name: "set string to interface field",
			targetStructFn: func() any {
				return &struct{ Value any }{}
			},
			fieldName: "Value",
			value:     "interface value",
			expected:  "interface value",
		},

		// Time tests
		{
			name: "set time.Time to time.Time field",
			targetStructFn: func() any {
				return &struct{ Value time.Time }{}
			},
			fieldName: "Value",
			value:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},

		// Complex tests
		{
			name: "set complex128 to complex128 field",
			targetStructFn: func() any {
				return &struct{ Value complex128 }{}
			},
			fieldName: "Value",
			value:     complex(1, 2),
			expected:  complex(1, 2),
		},

		// Map to map tests
		{
			name: "set map[string]any to map[string]any field",
			targetStructFn: func() any {
				return &struct{ Value map[string]any }{}
			},
			fieldName: "Value",
			value:     map[string]any{"key": "value", "num": 123},
			expected:  map[string]any{"key": "value", "num": 123},
		},

		// Pointer to slice
		{
			name: "set *[]string to []string field",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName: "Value",
			value:     &[]string{"a", "b"},
			expected:  []string{"a", "b"},
		},

		// Pointer to map
		{
			name: "set *map[string]string to map[string]string field",
			targetStructFn: func() any {
				return &struct{ Value map[string]string }{}
			},
			fieldName: "Value",
			value:     &map[string]string{"a": "b"},
			expected:  map[string]string{"a": "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			// Call Set
			err := Set(field, tc.value)

			// Assert
			require.NoError(t, err)

			// Get the actual value to compare
			var actual any
			if field.Kind() == reflect.Pointer && !field.IsNil() {
				actual = field.Elem().Interface()
			} else {
				actual = field.Interface()
			}

			// For nil pointer fields
			if field.Kind() == reflect.Pointer && field.IsNil() {
				require.Equal(t, tc.expected, actual)
			} else {
				switch actualValue := actual.(type) {
				case []string:
					if expectedValue, ok := tc.expected.([]string); ok {
						require.Equal(t, expectedValue, actualValue)
					} else {
						require.Equal(t, tc.expected, actual)
					}
				case []int:
					if expectedValue, ok := tc.expected.([]int); ok {
						require.Equal(t, expectedValue, actualValue)
					} else {
						require.Equal(t, tc.expected, actual)
					}
				case map[string]string:
					if expectedValue, ok := tc.expected.(map[string]string); ok {
						require.Equal(t, expectedValue, actualValue)
					} else {
						require.Equal(t, tc.expected, actual)
					}
				default:
					require.Equal(t, tc.expected, actual)
				}
			}
		})
	}
}

// Tests for failed value setting
func TestSet_Failure(t *testing.T) {
	// Define test cases for failures
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expectedError  error
	}{
		{
			name: "setting string to non-settable field",
			targetStructFn: func() any {
				// Create a struct type dynamically during runtime
				// with an unexported field that won't be settable
				type testStruct struct {
					value string // unexported field
				}
				return &testStruct{}
			},
			fieldName:     "value", // private field, not settable
			value:         "test",
			expectedError: errors.New("field of type string is not settable"),
		},
		{
			name: "setting int to struct field",
			targetStructFn: func() any {
				return &struct{ Value struct{} }{}
			},
			fieldName:     "Value",
			value:         42,
			expectedError: rerr.NotSupported,
		},
		{
			name: "setting []any to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName:     "Value",
			value:         []any{1, 2, 3},
			expectedError: rerr.NotSupported,
		},
		{
			name: "setting negative int to uint field",
			targetStructFn: func() any {
				return &struct{ Value uint }{}
			},
			fieldName:     "Value",
			value:         -42,
			expectedError: errors.New("cannot set negative value -42 to unsigned type uint"),
		},
		{
			name: "setting value that overflows target type",
			targetStructFn: func() any {
				return &struct{ Value int8 }{}
			},
			fieldName:     "Value",
			value:         1000, // too large for int8
			expectedError: errors.New("value 1000 is outside the range of target type int8"),
		},
		{
			name: "setting []any with incompatible elements to []string",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName:     "Value",
			value:         []any{1, true, struct{}{}}, // Can't convert to []string
			expectedError: errors.New("failed to convert element of []any to string"),
		},

		// Unsupported type tests
		{
			name: "setting func to int field",
			targetStructFn: func() any {
				return &struct{ Value int }{}
			},
			fieldName:     "Value",
			value:         func() {},
			expectedError: rerr.NotSupported,
		},
		{
			name: "setting chan to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName:     "Value",
			value:         make(chan int),
			expectedError: rerr.NotSupported,
		},

		// Map to incompatible map
		{
			name: "setting map[string]any to map[string]int field",
			targetStructFn: func() any {
				return &struct{ Value map[string]int }{}
			},
			fieldName:     "Value",
			value:         map[string]any{"key": "value"}, // value is string, not int
			expectedError: rerr.NotSupported,              // This should fail because the value types are incompatible
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()

			// Access field using reflect
			field := reflect.ValueOf(target).Elem().FieldByName(tc.fieldName)

			// Call Set
			err := Set(field, tc.value)

			// Assert error
			require.Error(t, err)
			if tc.expectedError != nil {
				// Use errors.Is for proper error comparison
				require.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"Expected error %v, got: %v", tc.expectedError, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkSet(b *testing.B) {
	// Create benchmark cases
	benchCases := []struct {
		name      string
		setupFn   func() (reflect.Value, any)
		valueType string
	}{
		{
			name: "string to string",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value string }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := "benchmark string value"
				return field, value
			},
			valueType: "string",
		},
		{
			name: "int to int",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value int }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := 42
				return field, value
			},
			valueType: "int",
		},
		{
			name: "float64 to float64",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value float64 }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := 3.14159
				return field, value
			},
			valueType: "float64",
		},
		{
			name: "bool to bool",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value bool }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := true
				return field, value
			},
			valueType: "bool",
		},
		{
			name: "string to *string",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value *string }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := "benchmark string pointer value"
				return field, value
			},
			valueType: "string to *string",
		},
		{
			name: "[]string to []string",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value []string }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := []string{"one", "two", "three"}
				return field, value
			},
			valueType: "[]string",
		},
		{
			name: "[]any to []string",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value []string }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := []any{"one", "two", "three"}
				return field, value
			},
			valueType: "[]any",
		},
		{
			name: "fmt.Stringer to string",
			setupFn: func() (reflect.Value, any) {
				s := struct{ Value string }{}
				field := reflect.ValueOf(&s).Elem().FieldByName("Value")
				value := CustomStringer{Value: "stringer benchmark value"}
				return field, value
			},
			valueType: "fmt.Stringer",
		},
	}

	// Run benchmarks
	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			field, value := bc.setupFn()

			// Reset timer before the loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = Set(field, value)
			}
		})
	}
}

// BenchmarkSet_Mixed measures performance when setting values of different types
func BenchmarkSet_Mixed(b *testing.B) {
	// Create a struct with various field types
	type MixedStruct struct {
		StringField    string
		IntField       int
		FloatField     float64
		BoolField      bool
		PtrField       *string
		SliceField     []string
		InterfaceField any
	}

	// Prepare values to set
	s := "string value"
	stringVal := "test string"
	intVal := 42
	floatVal := 3.14159
	boolVal := true
	ptrVal := &s
	sliceVal := []string{"one", "two", "three"}
	interfaceVal := CustomStringer{Value: "stringer value"}

	// Setup benchmark
	mixed := MixedStruct{}
	mixedValue := reflect.ValueOf(&mixed).Elem()
	fields := []struct {
		name  string
		value any
	}{
		{"StringField", stringVal},
		{"IntField", intVal},
		{"FloatField", floatVal},
		{"BoolField", boolVal},
		{"PtrField", ptrVal},
		{"SliceField", sliceVal},
		{"InterfaceField", interfaceVal},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// For each iteration, set all fields
		for _, field := range fields {
			fieldValue := mixedValue.FieldByName(field.name)
			_ = Set(fieldValue, field.value)
		}
	}
}
