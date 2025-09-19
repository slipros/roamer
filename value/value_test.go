package value

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
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

// TestSet_EdgeCases tests edge cases for the Set function
func TestSet_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
		expectError    bool
		errorContains  string
	}{
		// Test empty string handling
		{
			name: "empty string to various types",
			targetStructFn: func() any {
				return &struct {
					Str   string
					Int   int
					Bool  bool
					Float float64
					Slice []string
				}{
					// Pre-fill with non-zero values to test empty string behavior
					Str:   "not empty",
					Int:   42,
					Bool:  true,
					Float: 3.14,
					Slice: []string{"existing"},
				}
			},
			fieldName: "Str",
			value:     "",
			expected:  "",
		},
		// Test large slice handling
		{
			name: "large slice to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     make([]string, 1000),     // Large slice to test performance
			expected:  strings.Repeat(",", 999), // 1000 empty strings joined by commas
		},
		// Test deeply nested pointer types
		{
			name: "string to **string field",
			targetStructFn: func() any {
				return &struct{ Value **string }{}
			},
			fieldName: "Value",
			value:     "nested pointer test",
			expected:  "nested pointer test",
		},
		// Test nil pointer dereference safety
		{
			name: "nil pointer to value field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     (*string)(nil),
			expected:  "",
		},
		// Test any with various types
		{
			name: "complex type to interface field",
			targetStructFn: func() any {
				return &struct{ Value any }{}
			},
			fieldName: "Value",
			value:     map[string]any{"key": "value", "nested": map[string]string{"inner": "data"}},
			expected:  map[string]any{"key": "value", "nested": map[string]string{"inner": "data"}},
		},
		// Test channel type (should fail)
		{
			name: "channel type to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName:     "Value",
			value:         make(chan string, 1),
			expectError:   true,
			errorContains: "not supported",
		},
		// Test function type (any accepts anything, so this succeeds)
		{
			name: "function type to interface field",
			targetStructFn: func() any {
				return &struct{ Value any }{}
			},
			fieldName:   "Value",
			value:       func() string { return "test" },
			expectError: false,
			expected:    "function", // We'll check the type in a special way
		},
		// Test zero values
		{
			name: "zero int to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     0,
			expected:  "0",
		},
		// Test boundary values for integer types
		{
			name: "max int64 to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     int64(9223372036854775807),
			expected:  "9223372036854775807",
		},
		// Test negative zero float
		{
			name: "negative zero float to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     math.Copysign(0, -1),
			expected:  "-0",
		},
		// Test NaN and Infinity
		{
			name: "NaN to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     math.NaN(),
			expected:  "NaN",
		},
		{
			name: "positive infinity to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     math.Inf(1),
			expected:  "+Inf",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			// Call Set
			err := Set(field, tc.value)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errorContains))
				}
				return
			}

			// Assert no error for success cases
			require.NoError(t, err)

			// Get actual value
			actual := field.Interface()
			if field.Kind() == reflect.Pointer && !field.IsNil() {
				// Dereference pointer to get the actual value for comparison
				ptrVal := field
				for ptrVal.Kind() == reflect.Pointer && !ptrVal.IsNil() {
					ptrVal = ptrVal.Elem()
				}
				if ptrVal.Kind() != reflect.Pointer {
					actual = ptrVal.Interface()
				}
			}

			// Special handling for function comparison
			if tc.expected == "function" {
				// Check that the actual value is a function
				actualValue := reflect.ValueOf(actual)
				require.Equal(t, reflect.Func, actualValue.Kind(), "Expected function type")
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// TestSet_ConcurrentAccess tests concurrent access to Set function
func TestSet_ConcurrentAccess(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 100

	// Test concurrent setting of different fields - each goroutine gets its own struct
	t.Run("concurrent different fields", func(t *testing.T) {
		type ConcurrentStruct struct {
			Field1 string
			Field2 int
			Field3 bool
			Field4 []string
		}

		var wg sync.WaitGroup
		errorCh := make(chan error, numGoroutines*numIterations)

		// Launch goroutines to set fields on separate instances to avoid race conditions
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// Each goroutine gets its own struct instance to avoid races
				target := &ConcurrentStruct{}
				targetValue := reflect.ValueOf(target).Elem()

				for j := 0; j < numIterations; j++ {
					fieldName := fmt.Sprintf("Field%d", (id%4)+1)
					field := targetValue.FieldByName(fieldName)

					var value any
					switch fieldName {
					case "Field1":
						value = fmt.Sprintf("string-%d-%d", id, j)
					case "Field2":
						value = id*1000 + j
					case "Field3":
						value = (id+j)%2 == 0
					case "Field4":
						value = []string{fmt.Sprintf("item-%d", id), fmt.Sprintf("item-%d", j)}
					}

					if err := Set(field, value); err != nil {
						errorCh <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, err)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errorCh)

		// Check for errors
		for err := range errorCh {
			t.Errorf("Concurrent access error: %v", err)
		}
	})
}

// TestSet_MemoryUsage tests memory allocation patterns
func TestSet_MemoryUsage(t *testing.T) {
	tests := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		maxAllocs      int // Maximum expected allocations
	}{
		{
			name: "string to string (should be minimal allocs)",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     "test string",
			maxAllocs: 2, // Conservative estimate
		},
		{
			name: "int to int (should be zero allocs)",
			targetStructFn: func() any {
				return &struct{ Value int }{}
			},
			fieldName: "Value",
			value:     42,
			maxAllocs: 0,
		},
		{
			name: "slice to slice (some allocs expected)",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName: "Value",
			value:     []string{"one", "two", "three"},
			maxAllocs: 10, // More lenient for slice operations
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			// Warm up
			for i := 0; i < 10; i++ {
				_ = Set(field, tc.value)
			}

			// Measure allocations
			allocs := testing.AllocsPerRun(100, func() {
				_ = Set(field, tc.value)
			})

			if allocs > float64(tc.maxAllocs) {
				t.Errorf("Too many allocations: got %.2f, max %d", allocs, tc.maxAllocs)
			} else {
				t.Logf("Allocations: %.2f (max %d)", allocs, tc.maxAllocs)
			}
		})
	}
}

// TestHandleInterfaceSlice_Successfully tests successful conversion scenarios for handleInterfaceSlice
func TestHandleInterfaceSlice_Successfully(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          []any
		expected       any
	}{
		{
			name: "[]any to []string with strings",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName: "Value",
			value:     []any{"one", "two", "three"},
			expected:  []string{"one", "two", "three"},
		},
		{
			name: "[]any to []int with integers",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName: "Value",
			value:     []any{1, 2, 3},
			expected:  []int{1, 2, 3},
		},
		{
			name: "[]any to []int64 with mixed integer types",
			targetStructFn: func() any {
				return &struct{ Value []int64 }{}
			},
			fieldName: "Value",
			value:     []any{int8(1), int16(2), int32(3), int64(4)},
			expected:  []int64{1, 2, 3, 4},
		},
		{
			name: "[]any to []float64 with mixed numeric types",
			targetStructFn: func() any {
				return &struct{ Value []float64 }{}
			},
			fieldName: "Value",
			value:     []any{1, 2.5, float32(3.14), float64(4.0)},
			expected:  []float64{1.0, 2.5, float64(float32(3.14)), 4.0},
		},
		{
			name: "[]any to []bool with boolean values",
			targetStructFn: func() any {
				return &struct{ Value []bool }{}
			},
			fieldName: "Value",
			value:     []any{true, false, true},
			expected:  []bool{true, false, true},
		},
		{
			name: "empty []any to []string",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName: "Value",
			value:     []any{},
			expected:  []string{},
		},
		{
			name: "[]any to []any with mixed types",
			targetStructFn: func() any {
				return &struct{ Value []any }{}
			},
			fieldName: "Value",
			value:     []any{"string", "number", "boolean", "float"},
			expected:  []any{"string", "number", "boolean", "float"},
		},
		{
			name: "[]any to []*string with string pointers",
			targetStructFn: func() any {
				return &struct{ Value []*string }{}
			},
			fieldName: "Value",
			value:     []any{"one", "two", "three"},
			expected: func() []*string {
				s1, s2, s3 := "one", "two", "three"
				return []*string{&s1, &s2, &s3}
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			// Call Set which should internally use handleInterfaceSlice
			err := Set(field, tc.value)

			// Assert no error
			require.NoError(t, err)

			// Get actual value
			actual := field.Interface()

			// Compare based on type
			switch expectedVal := tc.expected.(type) {
			case []string:
				actualVal, ok := actual.([]string)
				require.True(t, ok, "Expected []string, got %T", actual)
				require.Equal(t, expectedVal, actualVal)
			case []int:
				actualVal, ok := actual.([]int)
				require.True(t, ok, "Expected []int, got %T", actual)
				require.Equal(t, expectedVal, actualVal)
			case []int64:
				actualVal, ok := actual.([]int64)
				require.True(t, ok, "Expected []int64, got %T", actual)
				require.Equal(t, expectedVal, actualVal)
			case []float64:
				actualVal, ok := actual.([]float64)
				require.True(t, ok, "Expected []float64, got %T", actual)
				require.Equal(t, expectedVal, actualVal)
			case []bool:
				actualVal, ok := actual.([]bool)
				require.True(t, ok, "Expected []bool, got %T", actual)
				require.Equal(t, expectedVal, actualVal)
			case []any:
				actualVal, ok := actual.([]any)
				require.True(t, ok, "Expected []any, got %T", actual)
				require.Equal(t, expectedVal, actualVal)
			case []*string:
				actualVal, ok := actual.([]*string)
				require.True(t, ok, "Expected []*string, got %T", actual)
				require.Equal(t, len(expectedVal), len(actualVal))
				for i, exp := range expectedVal {
					require.Equal(t, *exp, *actualVal[i])
				}
			}
		})
	}
}

// TestHandleInterfaceSlice_Failure tests error scenarios for handleInterfaceSlice
func TestHandleInterfaceSlice_Failure(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          []any
		expectedError  error
	}{
		{
			name: "[]any to non-slice field should fail",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName:     "Value",
			value:         []any{"one", "two"},
			expectedError: rerr.NotSupported,
		},
		{
			name: "[]any with incompatible element types",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName:     "Value",
			value:         []any{"not", "a", "number"},
			expectedError: errors.New("failed to convert element"),
		},
		{
			name: "[]any with mixed incompatible types to []int",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName:     "Value",
			value:         []any{1, "string", 3.14, true},
			expectedError: errors.New("failed to convert element"),
		},
		{
			name: "[]any with function type to []string",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName:     "Value",
			value:         []any{func() {}, "string"},
			expectedError: errors.New("failed to convert element"),
		},
		{
			name: "[]any with channel type to []int",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName:     "Value",
			value:         []any{make(chan int), 42},
			expectedError: errors.New("failed to convert element"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			// Call Set which should internally use handleInterfaceSlice
			err := Set(field, tc.value)

			// Assert error occurred
			require.Error(t, err)
			if tc.expectedError != nil {
				// Use errors.Is for proper error comparison or check if error message contains expected text
				require.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"Expected error %v, got: %v", tc.expectedError, err)
			}
		})
	}
}

// TestSet_TypeConversionPanics tests panic recovery during type conversion
func TestSet_TypeConversionPanics(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expectError    bool
	}{
		{
			name: "overflow during int conversion should return error",
			targetStructFn: func() any {
				return &struct{ Value int8 }{}
			},
			fieldName:   "Value",
			value:       int64(300), // Overflows int8 max value (127)
			expectError: true,       // Library correctly validates ranges
		},
		{
			name: "underflow during uint conversion should return error",
			targetStructFn: func() any {
				return &struct{ Value uint8 }{}
			},
			fieldName:   "Value",
			value:       int(-1), // Cannot convert negative to uint
			expectError: true,    // Library correctly validates negative values for unsigned types
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			// This should not panic, even with problematic conversions
			err := Set(field, tc.value)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSet_DeepPointerChains tests handling of deeply nested pointer types
func TestSet_DeepPointerChains(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
	}{
		{
			name: "string to ***string (triple pointer)",
			targetStructFn: func() any {
				return &struct{ Value ***string }{}
			},
			fieldName: "Value",
			value:     "deep pointer test",
			expected:  "deep pointer test",
		},
		{
			name: "int to **int (double pointer)",
			targetStructFn: func() any {
				return &struct{ Value **int }{}
			},
			fieldName: "Value",
			value:     42,
			expected:  42,
		},
		{
			name: "nil to ****string (quad pointer)",
			targetStructFn: func() any {
				return &struct{ Value ****string }{}
			},
			fieldName: "Value",
			value:     nil,
			expected:  (****string)(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			err := Set(field, tc.value)
			require.NoError(t, err)

			// Get actual value by dereferencing pointers
			actual := field.Interface()
			if field.Kind() == reflect.Pointer && !field.IsNil() {
				// Dereference all pointer levels to get to the actual value
				ptrVal := field
				for ptrVal.Kind() == reflect.Pointer && !ptrVal.IsNil() {
					ptrVal = ptrVal.Elem()
				}
				if ptrVal.Kind() != reflect.Pointer {
					actual = ptrVal.Interface()
				}
			}

			// Special handling for nil values
			if tc.value == nil {
				require.Equal(t, tc.expected, field.Interface())
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// TestSet_SpecialNumericValues tests handling of special numeric values
func TestSet_SpecialNumericValues(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
	}{
		{
			name: "positive infinity float64",
			targetStructFn: func() any {
				return &struct{ Value float64 }{}
			},
			fieldName: "Value",
			value:     math.Inf(1),
			expected:  math.Inf(1),
		},
		{
			name: "negative infinity float64",
			targetStructFn: func() any {
				return &struct{ Value float64 }{}
			},
			fieldName: "Value",
			value:     math.Inf(-1),
			expected:  math.Inf(-1),
		},
		{
			name: "NaN float64",
			targetStructFn: func() any {
				return &struct{ Value float64 }{}
			},
			fieldName: "Value",
			value:     math.NaN(),
			expected:  "NaN", // We'll check with IsNaN
		},
		{
			name: "max int64",
			targetStructFn: func() any {
				return &struct{ Value int64 }{}
			},
			fieldName: "Value",
			value:     int64(math.MaxInt64),
			expected:  int64(math.MaxInt64),
		},
		{
			name: "min int64",
			targetStructFn: func() any {
				return &struct{ Value int64 }{}
			},
			fieldName: "Value",
			value:     int64(math.MinInt64),
			expected:  int64(math.MinInt64),
		},
		{
			name: "max uint64",
			targetStructFn: func() any {
				return &struct{ Value uint64 }{}
			},
			fieldName: "Value",
			value:     uint64(math.MaxUint64),
			expected:  uint64(math.MaxUint64),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			err := Set(field, tc.value)
			require.NoError(t, err)

			actual := field.Interface()

			// Special handling for NaN
			if tc.expected == "NaN" {
				actualFloat, ok := actual.(float64)
				require.True(t, ok)
				require.True(t, math.IsNaN(actualFloat))
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// TestSet_UnsettableFields tests handling of unsettable fields
func TestSet_UnsettableFields(t *testing.T) {
	t.Run("unexported field should fail", func(t *testing.T) {
		// Create a struct with unexported field
		type testStruct struct {
			value string // unexported field - cannot be set
		}

		target := &testStruct{}
		targetValue := reflect.ValueOf(target).Elem()
		field := targetValue.FieldByName("value")

		err := Set(field, "test value")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not settable")
	})

	t.Run("read-only reflect.Value should fail", func(t *testing.T) {
		// Create a reflect.Value that cannot be set
		str := "test"
		value := reflect.ValueOf(str) // This creates a non-settable reflect.Value

		err := Set(value, "new value")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not settable")
	})
}

// TestSet_ComplexDataTypes tests setting complex data types
func TestSet_ComplexDataTypes(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
		expectError    bool
	}{
		{
			name: "nested struct assignable to any",
			targetStructFn: func() any {
				return &struct{ Value any }{}
			},
			fieldName: "Value",
			value: struct {
				Name string
				Age  int
			}{Name: "John", Age: 30},
			expected: struct {
				Name string
				Age  int
			}{Name: "John", Age: 30},
			expectError: false,
		},
		{
			name: "slice of structs to any",
			targetStructFn: func() any {
				return &struct{ Value any }{}
			},
			fieldName: "Value",
			value: []struct {
				ID   int
				Name string
			}{{1, "Alice"}, {2, "Bob"}},
			expected: []struct {
				ID   int
				Name string
			}{{1, "Alice"}, {2, "Bob"}},
			expectError: false,
		},
		{
			name: "map with any values",
			targetStructFn: func() any {
				return &struct{ Value map[string]any }{}
			},
			fieldName:   "Value",
			value:       map[string]any{"name": "Alice", "age": 25, "active": true},
			expected:    map[string]any{"name": "Alice", "age": 25, "active": true},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			err := Set(field, tc.value)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				actual := field.Interface()
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// TestSet_PointerTypeConversions tests various pointer type conversions
func TestSet_PointerTypeConversions(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
	}{
		{
			name: "nil pointer of struct type to struct pointer",
			targetStructFn: func() any {
				type TestStruct struct {
					Value string
				}
				return &struct{ Value *TestStruct }{}
			},
			fieldName: "Value",
			value:     (*struct{ Value string })(nil),
			expected:  "zero_value", // Special marker to test differently
		},
		{
			name: "pointer to value with same underlying type",
			targetStructFn: func() any {
				return &struct{ Value *int }{}
			},
			fieldName: "Value",
			value: func() *int {
				val := 42
				return &val
			}(),
			expected: 42,
		},
		{
			name: "pointer to different integer type (should convert)",
			targetStructFn: func() any {
				return &struct{ Value int64 }{}
			},
			fieldName: "Value",
			value: func() *int32 {
				val := int32(42)
				return &val
			}(),
			expected: int64(42),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			err := Set(field, tc.value)
			require.NoError(t, err)

			// Handle pointer dereferencing for comparison
			actual := field.Interface()
			if field.Kind() == reflect.Pointer && !field.IsNil() {
				actual = field.Elem().Interface()
			}

			// Special handling for nil pointer values
			if tc.expected == "zero_value" {
				// For nil input values, Set function sets zero value, not nil
				if field.Kind() == reflect.Pointer {
					require.False(t, field.IsNil(), "Expected field to be initialized with zero value")
				}
			} else if tc.value == nil || reflect.ValueOf(tc.value).IsNil() {
				if field.Kind() == reflect.Pointer {
					require.True(t, field.IsNil(), "Expected field to be nil")
				} else {
					require.Equal(t, tc.expected, actual)
				}
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// TestSet_SliceVariations tests various slice type scenarios
func TestSet_SliceVariations(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          any
		expected       any
		expectError    bool
	}{
		{
			name: "nil slice to slice field",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName:   "Value",
			value:       ([]string)(nil),
			expected:    []string{}, // Set function initializes nil slices to empty slices
			expectError: false,
		},
		{
			name: "empty slice to slice field",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName:   "Value",
			value:       []int{},
			expected:    []int{},
			expectError: false,
		},
		{
			name: "slice with single element",
			targetStructFn: func() any {
				return &struct{ Value []string }{}
			},
			fieldName:   "Value",
			value:       []string{"single"},
			expected:    []string{"single"},
			expectError: false,
		},
		{
			name: "large slice (1000 elements)",
			targetStructFn: func() any {
				return &struct{ Value []int }{}
			},
			fieldName: "Value",
			value: func() []int {
				slice := make([]int, 1000)
				for i := range slice {
					slice[i] = i
				}
				return slice
			}(),
			expected: func() []int {
				slice := make([]int, 1000)
				for i := range slice {
					slice[i] = i
				}
				return slice
			}(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			err := Set(field, tc.value)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				actual := field.Interface()
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// StringerInt is a test type that implements fmt.Stringer
type StringerInt int

func (si StringerInt) String() string {
	return fmt.Sprintf("StringerInt(%d)", int(si))
}

// StringerStruct is a test type that implements fmt.Stringer
type StringerStruct struct {
	Value string
}

func (ss StringerStruct) String() string {
	return fmt.Sprintf("StringerStruct{Value: %s}", ss.Value)
}

// TestSet_StringerImplementations tests various types that implement fmt.Stringer
func TestSet_StringerImplementations(t *testing.T) {
	testCases := []struct {
		name           string
		targetStructFn func() any
		fieldName      string
		value          fmt.Stringer
		expected       string
	}{
		{
			name: "custom int stringer to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     StringerInt(42),
			expected:  "StringerInt(42)",
		},
		{
			name: "custom struct stringer to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     StringerStruct{Value: "test"},
			expected:  "StringerStruct{Value: test}",
		},
		{
			name: "time.Time stringer to string field",
			targetStructFn: func() any {
				return &struct{ Value string }{}
			},
			fieldName: "Value",
			value:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected:  "2023-01-01 12:00:00 +0000 UTC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.targetStructFn()
			targetValue := reflect.ValueOf(target).Elem()
			field := targetValue.FieldByName(tc.fieldName)

			err := Set(field, tc.value)
			require.NoError(t, err)

			actual := field.Interface()
			require.Equal(t, tc.expected, actual)
		})
	}
}

// TestSet_RaceConditionSafety tests thread safety more comprehensively
func TestSet_RaceConditionSafety(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 200

	t.Run("concurrent setting same field different values", func(t *testing.T) {
		type ConcurrentStruct struct {
			Value string
		}

		var wg sync.WaitGroup
		errorCh := make(chan error, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					target := &ConcurrentStruct{}
					targetValue := reflect.ValueOf(target).Elem()
					field := targetValue.FieldByName("Value")

					value := fmt.Sprintf("goroutine-%d-iteration-%d", id, j)

					if err := Set(field, value); err != nil {
						errorCh <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, err)
					}

					// Verify the value was set correctly
					if actual := field.Interface(); actual != value {
						errorCh <- fmt.Errorf("goroutine %d iteration %d: expected %s, got %s", id, j, value, actual)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errorCh)

		// Check for any errors
		for err := range errorCh {
			t.Errorf("Race condition error: %v", err)
		}
	})

	t.Run("concurrent setting mixed types", func(t *testing.T) {
		type MixedStruct struct {
			StringField string
			IntField    int
			BoolField   bool
			SliceField  []string
		}

		var wg sync.WaitGroup
		errorCh := make(chan error, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					target := &MixedStruct{}
					targetValue := reflect.ValueOf(target).Elem()

					// Set different fields with different types
					fields := []struct {
						name  string
						value any
					}{
						{"StringField", fmt.Sprintf("str-%d-%d", id, j)},
						{"IntField", id*1000 + j},
						{"BoolField", (id+j)%2 == 0},
						{"SliceField", []string{fmt.Sprintf("item-%d", id), fmt.Sprintf("item-%d", j)}},
					}

					for _, fieldInfo := range fields {
						field := targetValue.FieldByName(fieldInfo.name)
						if err := Set(field, fieldInfo.value); err != nil {
							errorCh <- fmt.Errorf("goroutine %d iteration %d field %s: %w", id, j, fieldInfo.name, err)
						}
					}
				}
			}(i)
		}

		wg.Wait()
		close(errorCh)

		// Check for any errors
		for err := range errorCh {
			t.Errorf("Mixed types race condition error: %v", err)
		}
	})
}

// TestSet_LargeDataHandling tests handling of large data structures
func TestSet_LargeDataHandling(t *testing.T) {
	t.Run("large string (1MB)", func(t *testing.T) {
		target := &struct{ Value string }{}
		field := reflect.ValueOf(target).Elem().FieldByName("Value")

		// Create a 1MB string
		largeString := strings.Repeat("a", 1024*1024)

		err := Set(field, largeString)
		require.NoError(t, err)
		require.Equal(t, largeString, field.Interface())
		require.Equal(t, 1024*1024, len(field.Interface().(string)))
	})

	t.Run("large slice (100k elements)", func(t *testing.T) {
		target := &struct{ Value []int }{}
		field := reflect.ValueOf(target).Elem().FieldByName("Value")

		// Create a slice with 100k elements
		largeSlice := make([]int, 100000)
		for i := range largeSlice {
			largeSlice[i] = i
		}

		err := Set(field, largeSlice)
		require.NoError(t, err)

		actual := field.Interface().([]int)
		require.Equal(t, len(largeSlice), len(actual))
		require.Equal(t, largeSlice, actual)
	})

	t.Run("deeply nested map", func(t *testing.T) {
		target := &struct{ Value any }{}
		field := reflect.ValueOf(target).Elem().FieldByName("Value")

		// Create a deeply nested map structure
		deepMap := make(map[string]any)
		current := deepMap

		// Create 10 levels of nesting
		for i := 0; i < 10; i++ {
			next := make(map[string]any)
			current[fmt.Sprintf("level_%d", i)] = next
			current = next
		}
		current["final_value"] = "deep_nested_value"

		err := Set(field, deepMap)
		require.NoError(t, err)
		require.Equal(t, deepMap, field.Interface())
	})
}
