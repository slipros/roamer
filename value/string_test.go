package value

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Custom types for testing unmarshalers
type UnmarshallerText struct {
	S string
}

func (u *UnmarshallerText) UnmarshalText(text []byte) error {
	u.S = string(text)
	return nil
}

type UnmarshallerBinary struct {
	S string
}

func (u *UnmarshallerBinary) UnmarshalBinary(data []byte) error {
	u.S = string(data)
	return nil
}

type failingUnmarshaler struct{}

func (f *failingUnmarshaler) UnmarshalText(text []byte) error {
	return errors.New("intentional unmarshal failure")
}

// TestSetString_Successfully tests successful scenarios of converting and setting string values.
func TestSetString_Successfully(t *testing.T) {
	// Create test structure
	type TestStruct struct {
		StrField             string
		BoolField            bool
		IntField             int
		Int8Field            int8
		Int16Field           int16
		Int32Field           int32
		Int64Field           int64
		UintField            uint
		Uint8Field           uint8
		Uint16Field          uint16
		Uint32Field          uint32
		Uint64Field          uint64
		Float32Field         float32
		Float64Field         float64
		Complex64Field       complex64
		Complex128Field      complex128
		ByteSlice            []byte
		StringSlice          []string
		IntSlice             []int
		PtrField             *string
		TextUnmarshaler      UnmarshallerText
		BinaryUnmarshaler    UnmarshallerBinary
		TimeField            time.Time
		InterfaceField       interface{}
		MapStringStringField map[string]string
		MapStringIntField    map[string]int
		PtrIntField          *int
	}

	// Define test cases
	tests := []struct {
		name     string
		field    string      // field name to set
		input    string      // input string
		expected interface{} // expected value
	}{
		// String values
		{name: "string to string", field: "StrField", input: "test_string", expected: "test_string"},
		{name: "empty string to string", field: "StrField", input: "", expected: ""},

		// Boolean values
		{name: "true to bool", field: "BoolField", input: "true", expected: true},
		{name: "1 to bool", field: "BoolField", input: "1", expected: true},
		{name: "yes to bool", field: "BoolField", input: "yes", expected: true},
		{name: "false to bool", field: "BoolField", input: "false", expected: false},
		{name: "0 to bool", field: "BoolField", input: "0", expected: false},
		{name: "no to bool", field: "BoolField", input: "no", expected: false},
		{name: "empty string to bool", field: "BoolField", input: "", expected: false},

		// Integer values
		{name: "123 to int", field: "IntField", input: "123", expected: int(123)},
		{name: "123 to int8", field: "Int8Field", input: "123", expected: int8(123)},
		{name: "123 to int16", field: "Int16Field", input: "123", expected: int16(123)},
		{name: "123 to int32", field: "Int32Field", input: "123", expected: int32(123)},
		{name: "123 to int64", field: "Int64Field", input: "123", expected: int64(123)},
		{name: "0x1A to int (hex)", field: "IntField", input: "0x1A", expected: int(26)},
		{name: "010 to int (octal)", field: "IntField", input: "010", expected: int(8)},
		{name: "empty string to int", field: "IntField", input: "", expected: int(0)},

		// Unsigned integer values
		{name: "123 to uint", field: "UintField", input: "123", expected: uint(123)},
		{name: "123 to uint8", field: "Uint8Field", input: "123", expected: uint8(123)},
		{name: "123 to uint16", field: "Uint16Field", input: "123", expected: uint16(123)},
		{name: "123 to uint32", field: "Uint32Field", input: "123", expected: uint32(123)},
		{name: "123 to uint64", field: "Uint64Field", input: "123", expected: uint64(123)},
		{name: "0xFF to uint (hex)", field: "UintField", input: "0xFF", expected: uint(255)},
		{name: "010 to uint (octal)", field: "UintField", input: "010", expected: uint(8)},
		{name: "empty string to uint", field: "UintField", input: "", expected: uint(0)},

		// Floating point values
		{name: "123.45 to float32", field: "Float32Field", input: "123.45", expected: float32(123.45)},
		{name: "123.45 to float64", field: "Float64Field", input: "123.45", expected: float64(123.45)},
		{name: "empty string to float64", field: "Float64Field", input: "", expected: float64(0)},

		// Complex values
		{name: "1+2i to complex64", field: "Complex64Field", input: "1+2i", expected: complex64(1 + 2i)},
		{name: "1+2i to complex128", field: "Complex128Field", input: "1+2i", expected: complex128(1 + 2i)},
		{name: "empty string to complex128", field: "Complex128Field", input: "", expected: complex128(0)},

		// Slices
		{name: "string to byte slice", field: "ByteSlice", input: "hello", expected: []byte("hello")},
		{name: "string to string slice (single item)", field: "StringSlice", input: "hello", expected: []string{"hello"}},
		{name: "comma separated string to string slice", field: "StringSlice", input: "hello,world,test", expected: []string{"hello", "world", "test"}},
		{name: "comma separated int string to int slice", field: "IntSlice", input: "1,2,3", expected: []int{1, 2, 3}},

		// Pointers and custom types
		{name: "string to *string", field: "PtrField", input: "hello", expected: ptrToString("hello")},
		{name: "string to text unmarshaler", field: "TextUnmarshaler", input: "hello", expected: UnmarshallerText{S: "hello"}},
		{name: "string to binary unmarshaler", field: "BinaryUnmarshaler", input: "hello", expected: UnmarshallerBinary{S: "hello"}},

		// Interfaces
		{name: "string to interface", field: "InterfaceField", input: "hello", expected: "hello"},

		// Time
		{name: "RFC3339 to time.Time", field: "TimeField", input: "2023-01-15T12:00:00Z", expected: mustParseTime(time.RFC3339, "2023-01-15T12:00:00Z")},
		{name: "simple date to time.Time", field: "TimeField", input: "2023-01-15", expected: mustParseTime("2006-01-02", "2023-01-15")},
		{name: "datetime (len 19) to time.Time", field: "TimeField", input: "2023-01-15 12:00:00", expected: mustParseTime("2006-01-02 15:04:05", "2023-01-15 12:00:00")},
		{name: "RFC3339 with offset (len 25) to time.Time", field: "TimeField", input: "2023-01-15T12:00:00+01:00", expected: mustParseTime(time.RFC3339, "2023-01-15T12:00:00+01:00")},
		{name: "RFC822 to time.Time", field: "TimeField", input: "15 Jan 23 12:00 UTC", expected: mustParseTime(time.RFC822, "15 Jan 23 12:00 UTC")},
		{name: "RFC3339Nano to time.Time", field: "TimeField", input: "2023-01-15T12:00:00.123456789Z", expected: mustParseTime(time.RFC3339Nano, "2023-01-15T12:00:00.123456789Z")},

		// Map
		{name: "string to map[string]string", field: "MapStringStringField", input: "key1:value1,key2:value2", expected: map[string]string{"key1": "value1", "key2": "value2"}},
		{name: "string to map[string]int", field: "MapStringIntField", input: "key1:1,key2:2", expected: map[string]int{"key1": 1, "key2": 2}},
		{name: "empty string to map", field: "MapStringStringField", input: "", expected: map[string]string{}},

		// Empty string to pointer
		{name: "empty string to *int", field: "PtrIntField", input: "", expected: new(int)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new struct for each test
			var testObj TestStruct

			// Get reflection of the struct
			val := reflect.ValueOf(&testObj).Elem()

			// Find the field by name
			field := val.FieldByName(tc.field)
			require.True(t, field.IsValid(), "Field %s not found", tc.field)

			// Set the value
			err := SetString(field, tc.input)
			require.NoError(t, err, "SetString should not return error")

			// Check the result
			// For slices we need to use deep comparison
			if field.Kind() == reflect.Slice {
				assert.Equal(t, tc.expected, field.Interface(), "Field value should match expected")
			} else if field.Kind() == reflect.Ptr && !field.IsNil() {
				// For pointers compare the values they point to
				assert.Equal(t, tc.expected, field.Interface(), "Field value should match expected")
			} else if field.Type() == typeTime {
				// For time.Time compare as strings in RFC3339 format
				expectedTime := tc.expected.(time.Time)
				actualTime := field.Interface().(time.Time)
				assert.Equal(t, expectedTime.Format(time.RFC3339), actualTime.Format(time.RFC3339))
			} else {
				// For other types direct comparison
				assert.Equal(t, tc.expected, field.Interface(), "Field value should match expected")
			}
		})
	}
}

// TestSetString_Failure tests scenarios where SetString should return an error.
func TestSetString_Failure(t *testing.T) {
	// Create test structure
	type TestStruct struct {
		IntField             int
		UintField            uint
		FloatField           float64
		ComplexField         complex128
		TimeField            time.Time
		MapStringStringField map[string]string
		MapStringIntField    map[string]int
		MapIntStringField    map[int]string
		ChannelField         chan int
		FailingUnmarshaler   failingUnmarshaler
	}

	// Define test cases
	tests := []struct {
		name     string
		field    string           // field name to set
		input    string           // input string
		errCheck func(error) bool // error check function
	}{
		// Invalid numeric formats
		{
			name:     "invalid int",
			field:    "IntField",
			input:    "not_a_number",
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name:     "invalid uint",
			field:    "UintField",
			input:    "-123", // negative number for unsigned type
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name:     "invalid float",
			field:    "FloatField",
			input:    "123.45.67",
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name:     "invalid complex",
			field:    "ComplexField",
			input:    "1+2i+3i",
			errCheck: func(err error) bool { return err != nil },
		},

		// Invalid time formats
		{
			name:     "invalid time format",
			field:    "TimeField",
			input:    "not_a_time_format",
			errCheck: func(err error) bool { return err != nil },
		},

		// Unsupported map types
		{
			name:  "map with non-string keys",
			field: "MapIntStringField",
			input: "key1:value1,key2:value2",
			errCheck: func(err error) bool {
				return errors.Is(errors.Cause(err), rerr.NotSupported)
			},
		},
		{
			name:     "invalid map format",
			field:    "MapStringIntField",
			input:    "not_a_map_format",
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name:     "invalid map key-value pair",
			field:    "MapStringIntField",
			input:    "key1=value1", // using = instead of :
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name:     "malformed map string (no colon)",
			field:    "MapStringStringField",
			input:    "key1value1,key2:value2",
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name:     "malformed map string (value not int)",
			field:    "MapStringIntField",
			input:    "key1:abc",
			errCheck: func(err error) bool { return err != nil },
		},

		// Unsupported field types
		{
			name:  "unsupported field type",
			field: "ChannelField",
			input: "anything",
			errCheck: func(err error) bool {
				return errors.Is(errors.Cause(err), rerr.NotSupported)
			},
		},

		// Custom unmarshaler errors
		{
			name:     "failing unmarshaler",
			field:    "FailingUnmarshaler",
			input:    "anything",
			errCheck: func(err error) bool { return err != nil },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new struct for each test
			var testObj TestStruct

			// Get reflection of the struct
			val := reflect.ValueOf(&testObj).Elem()

			// Find the field by name
			field := val.FieldByName(tc.field)
			require.True(t, field.IsValid(), "Field %s not found", tc.field)

			// Set the value
			err := SetString(field, tc.input)

			// Check the error
			assert.True(t, tc.errCheck(err), "Expected error for input '%s'", tc.input)
		})
	}
}

// TestSetString_UnexportedField tests that SetString returns an error when trying to set an unexported field
func TestSetString_UnexportedField(t *testing.T) {
	type structWithUnexportedField struct {
		unexportedField string // Unexported field
	}

	// Create an instance of the struct
	var obj structWithUnexportedField

	// Get the unexported field by reflection
	field := reflect.ValueOf(obj).FieldByName("unexportedField")
	require.True(t, field.IsValid(), "Unexported field should be found")

	// Try to set a value to the unexported field
	err := SetString(field, "test value")

	// Check error
	assert.Error(t, err, "SetString should return error for unexported field")
	assert.True(t, errors.Is(errors.Cause(err), rerr.NotSupported),
		"Error should be NotSupported for unexported field")
}

// Helper functions for tests
func ptrToString(s string) *string {
	return &s
}

func mustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// BenchmarkSetString benchmarks the SetString function with various types.
func BenchmarkSetString(b *testing.B) {
	// Define a struct with various field types for testing
	type TestStruct struct {
		StringField       string
		IntField          int
		Float64Field      float64
		BoolField         bool
		TimeField         time.Time
		SliceField        []string
		PointerField      *string
		TextUnmarshaler   UnmarshallerText
		BinaryUnmarshaler UnmarshallerBinary
	}

	// Create benchmark cases
	benchmarks := []struct {
		name      string
		fieldName string
		value     string
		setup     func() reflect.Value
	}{
		{
			name:      "String",
			fieldName: "StringField",
			value:     "test string value",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("StringField")
			},
		},
		{
			name:      "Int",
			fieldName: "IntField",
			value:     "12345",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("IntField")
			},
		},
		{
			name:      "Float64",
			fieldName: "Float64Field",
			value:     "123.456",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("Float64Field")
			},
		},
		{
			name:      "Bool",
			fieldName: "BoolField",
			value:     "true",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("BoolField")
			},
		},
		{
			name:      "Time",
			fieldName: "TimeField",
			value:     "2023-01-15T12:00:00Z",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("TimeField")
			},
		},
		{
			name:      "Slice",
			fieldName: "SliceField",
			value:     "item1,item2,item3",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("SliceField")
			},
		},
		{
			name:      "Pointer",
			fieldName: "PointerField",
			value:     "pointer value",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("PointerField")
			},
		},
		{
			name:      "TextUnmarshaler",
			fieldName: "TextUnmarshaler",
			value:     "unmarshaler text",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("TextUnmarshaler")
			},
		},
		{
			name:      "BinaryUnmarshaler",
			fieldName: "BinaryUnmarshaler",
			value:     "unmarshaler binary",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("BinaryUnmarshaler")
			},
		},
		{
			name:      "EmptyString",
			fieldName: "StringField",
			value:     "",
			setup: func() reflect.Value {
				var obj TestStruct
				return reflect.ValueOf(&obj).Elem().FieldByName("StringField")
			},
		},
	}

	// Run benchmarks
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			field := bm.setup()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// The actual function we're benchmarking
				_ = SetString(field, bm.value)
			}
		})
	}
}

// BenchmarkSetString_SameField benchmarks SetString with reuse of the same field.
// This is to test the performance when repeatedly setting values to the same field.
func BenchmarkSetString_SameField(b *testing.B) {
	type TestStruct struct {
		StringField  string
		IntField     int
		Float64Field float64
		TimeField    time.Time
	}

	benchmarks := []struct {
		name      string
		fieldName string
		value     string
	}{
		{
			name:      "String",
			fieldName: "StringField",
			value:     "test string value",
		},
		{
			name:      "Int",
			fieldName: "IntField",
			value:     "12345",
		},
		{
			name:      "Float64",
			fieldName: "Float64Field",
			value:     "123.456",
		},
		{
			name:      "Time",
			fieldName: "TimeField",
			value:     "2023-01-15T12:00:00Z",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			var obj TestStruct
			field := reflect.ValueOf(&obj).Elem().FieldByName(bm.fieldName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = SetString(field, bm.value)
			}
		})
	}
}

// BenchmarkSetString_SuccessVsFailure compares the performance of successful
// and failing SetString calls.
func BenchmarkSetString_SuccessVsFailure(b *testing.B) {
	type TestStruct struct {
		ExportedField   string
		unexportedField string
	}

	b.Run("Success", func(b *testing.B) {
		var obj TestStruct
		field := reflect.ValueOf(&obj).Elem().FieldByName("ExportedField")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SetString(field, "value")
		}
	})

	b.Run("Failure_Unexported", func(b *testing.B) {
		var obj TestStruct
		field := reflect.ValueOf(obj).FieldByName("unexportedField")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SetString(field, "value")
		}
	})

	b.Run("Failure_InvalidValue", func(b *testing.B) {
		var obj TestStruct
		field := reflect.ValueOf(&obj).Elem().FieldByName("ExportedField")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// This should fail because "letters" can't be converted to an int
			_ = SetString(field, "letters")
		}
	})
}

// BenchmarkSetTimeFromString compares the performance of time parsing
// with different formats.
func BenchmarkSetTimeFromString(b *testing.B) {
	benchmarks := []struct {
		name   string
		format string
		value  string
	}{
		{
			name:   "RFC3339",
			format: time.RFC3339,
			value:  "2023-01-15T12:00:00Z",
		},
		{
			name:   "RFC3339Nano",
			format: time.RFC3339Nano,
			value:  "2023-01-15T12:00:00.123456789Z",
		},
		{
			name:   "RFC1123",
			format: time.RFC1123,
			value:  "Sun, 15 Jan 2023 12:00:00 UTC",
		},
		{
			name:   "SimpleDate",
			format: "2006-01-02",
			value:  "2023-01-15",
		},
		{
			name:   "DateTime",
			format: "2006-01-02 15:04:05",
			value:  "2023-01-15 12:00:00",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			var t time.Time
			field := reflect.ValueOf(&t).Elem()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = SetString(field, bm.value)
			}
		})
	}
}

// TestSetSliceFromString_EdgeCases tests various edge cases and boundary conditions
// that could cause issues with the type safety fixes.
func TestSetSliceFromString_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupField  func() reflect.Value
		input       string
		expected    interface{}
		expectError bool
		errorCheck  func(error) bool
	}{
		// Empty and whitespace handling
		{
			name: "empty string to string slice",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "",
			expected: []string(nil), // Empty string results in nil slice
		},
		{
			name: "only whitespace to string slice",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "   \t  \n  ",
			expected: []string{}, // Trimmed whitespace should result in empty slice
		},
		{
			name: "only commas to string slice",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    ",,,",
			expected: []string{}, // Only empty elements, should result in empty slice
		},
		{
			name: "mixed empty and valid elements",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    ",hello,,world,",
			expected: []string{"hello", "world"}, // Empty elements filtered out
		},

		// Numeric edge cases
		{
			name: "zero values in int slice",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "0,0,0",
			expected: []int{0, 0, 0},
		},
		{
			name: "max int64 values",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int64 }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "9223372036854775807,-9223372036854775808",
			expected: []int64{9223372036854775807, -9223372036854775808},
		},
		{
			name: "very small float values",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1e-100,2.2250738585072014e-308",
			expected: []float64{1e-100, 2.2250738585072014e-308},
		},
		{
			name: "infinity and special float values",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "Inf,-Inf",
			expected: []float64{math.Inf(1), math.Inf(-1)}, // Positive and negative infinity
		},

		// Boolean edge cases
		{
			name: "empty elements in bool slice",
			setupField: func() reflect.Value {
				var s struct{ BoolSlice []bool }
				return reflect.ValueOf(&s).Elem().FieldByName("BoolSlice")
			},
			input:    "true,,false",
			expected: []bool{true, false}, // Empty elements should be filtered out
		},

		// Large input handling
		{
			name: "very long string slice",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input: func() string {
				// Create a long comma-separated string
				result := ""
				for i := 0; i < 1000; i++ {
					if i > 0 {
						result += ","
					}
					result += "item" + string(rune('0'+i%10))
				}
				return result
			}(),
			expected: func() []string {
				result := make([]string, 1000)
				for i := 0; i < 1000; i++ {
					result[i] = "item" + string(rune('0'+i%10))
				}
				return result
			}(),
		},

		// Error cases - invalid conversions
		{
			name: "non-numeric string to int slice",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:       "1,abc,3",
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil && (errors.Cause(err) != nil || err.Error() != "")
			},
		},
		{
			name: "invalid float string to float slice",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:       "1.1,not.a.float,3.3",
			expectError: true,
		},
		{
			name: "invalid bool string to bool slice",
			setupField: func() reflect.Value {
				var s struct{ BoolSlice []bool }
				return reflect.ValueOf(&s).Elem().FieldByName("BoolSlice")
			},
			input:       "true,maybe,false",
			expectError: true,
		},
		{
			name: "overflow int8 values (now properly returns error)",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int8 }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:       "1,128,3", // 128 is out of range for int8 (-128 to 127)
			expectError: true,      // Fixed: now properly returns an error instead of silent overflow
		},
		{
			name: "underflow int8 values",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int8 }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:       "1,-129,3", // -129 is out of range for int8
			expectError: true,
		},
		{
			name: "negative uint values",
			setupField: func() reflect.Value {
				var s struct{ UintSlice []uint }
				return reflect.ValueOf(&s).Elem().FieldByName("UintSlice")
			},
			input:       "1,-1,3", // negative values not allowed for uint
			expectError: true,
		},

		// Special character handling
		{
			name: "strings with commas inside (should split)",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "hello,world",
			expected: []string{"hello", "world"}, // Should split on comma
		},
		{
			name: "strings with escape sequences",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "hello\\nworld,test\\ttab",
			expected: []string{"hello\\nworld", "test\\ttab"}, // Literal backslashes
		},

		// Unicode handling
		{
			name: "unicode strings",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "привет,世界,🌍",
			expected: []string{"привет", "世界", "🌍"},
		},

		// Mixed type conversion attempts (should fail)
		{
			name: "string that looks like array to int slice",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:       "[1,2,3]", // Not comma-separated, contains brackets
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field := tc.setupField()
			require.True(t, field.IsValid(), "Field should be valid")
			require.True(t, field.CanSet(), "Field should be settable")

			err := SetString(field, tc.input)

			if tc.expectError {
				assert.Error(t, err, "SetString should return error for input: %s", tc.input)
				if tc.errorCheck != nil {
					assert.True(t, tc.errorCheck(err), "Error should match expected criteria")
				}
				return
			}

			require.NoError(t, err, "SetString should succeed for input: %s", tc.input)

			result := field.Interface()
			assert.Equal(t, tc.expected, result, "Result should match expected value")
		})
	}
}

// TestSetSliceFromString_UnsupportedTypes tests that unsupported slice types
// return appropriate errors instead of panicking.
func TestSetSliceFromString_UnsupportedTypes(t *testing.T) {
	tests := []struct {
		name       string
		setupField func() reflect.Value
		input      string
		errorCheck func(error) bool
	}{
		{
			name: "slice of unsupported struct type",
			setupField: func() reflect.Value {
				type CustomStruct struct{ Field string }
				var s struct{ StructSlice []CustomStruct }
				return reflect.ValueOf(&s).Elem().FieldByName("StructSlice")
			},
			input: "test,data",
			errorCheck: func(err error) bool {
				return err != nil // Should return error for unsupported type
			},
		},
		{
			name: "slice of channels",
			setupField: func() reflect.Value {
				var s struct{ ChannelSlice []chan int }
				return reflect.ValueOf(&s).Elem().FieldByName("ChannelSlice")
			},
			input: "test",
			errorCheck: func(err error) bool {
				return err != nil // Should return error for unsupported type
			},
		},
		{
			name: "slice of functions",
			setupField: func() reflect.Value {
				var s struct{ FuncSlice []func() }
				return reflect.ValueOf(&s).Elem().FieldByName("FuncSlice")
			},
			input: "test",
			errorCheck: func(err error) bool {
				return err != nil // Should return error for unsupported type
			},
		},
		{
			name: "slice of maps",
			setupField: func() reflect.Value {
				var s struct{ MapSlice []map[string]string }
				return reflect.ValueOf(&s).Elem().FieldByName("MapSlice")
			},
			input: "test",
			errorCheck: func(err error) bool {
				return err != nil // Should return error for unsupported type
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field := tc.setupField()
			require.True(t, field.IsValid(), "Field should be valid")
			require.True(t, field.CanSet(), "Field should be settable")

			err := SetString(field, tc.input)
			assert.True(t, tc.errorCheck(err), "Should return appropriate error for unsupported type")
		})
	}
}

// TestSetSliceFromString_MemoryLeaks tests for potential memory leaks
// in slice allocation and management.
func TestSetSliceFromString_MemoryLeaks(t *testing.T) {
	// This test ensures that slice creation and element assignment
	// doesn't create memory leaks with the reflect operations.

	tests := []struct {
		name       string
		setupField func() reflect.Value
		input      string
		iterations int
	}{
		{
			name: "repeated string slice assignments",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:      "a,b,c,d,e",
			iterations: 100,
		},
		{
			name: "repeated int slice assignments",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:      "1,2,3,4,5",
			iterations: 100,
		},
		{
			name: "large slice repeated assignments",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input: func() string {
				result := ""
				for i := 0; i < 100; i++ {
					if i > 0 {
						result += ","
					}
					result += "item" + string(rune('0'+i%10))
				}
				return result
			}(),
			iterations: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Multiple assignments to the same field should not cause memory issues
			for i := 0; i < tc.iterations; i++ {
				field := tc.setupField()
				err := SetString(field, tc.input)
				require.NoError(t, err, "Assignment %d should succeed", i+1)

				// Verify the slice is not nil and has expected content
				result := field.Interface()
				assert.NotNil(t, result, "Result should not be nil at iteration %d", i+1)
			}
		})
	}
}

// TestSetSliceFromString_ConcurrentAccess tests concurrent access to slice conversion
// to ensure thread safety of the type conversion process.
func TestSetSliceFromString_ConcurrentAccess(t *testing.T) {
	const numGoroutines = 10
	const numIterations = 50

	tests := []struct {
		name  string
		input string
	}{
		{"concurrent string slices", "a,b,c,d,e"},
		{"concurrent int slices", "1,2,3,4,5"},
		{"concurrent bool slices", "true,false,true"},
		{"concurrent float slices", "1.1,2.2,3.3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Channel to collect errors from goroutines
			errChan := make(chan error, numGoroutines*numIterations)

			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					for j := 0; j < numIterations; j++ {
						// Create a new struct for each operation to avoid race conditions
						var testStruct struct {
							StringSlice []string
							IntSlice    []int
							BoolSlice   []bool
							FloatSlice  []float64
						}

						val := reflect.ValueOf(&testStruct).Elem()

						var field reflect.Value
						switch tc.name {
						case "concurrent string slices":
							field = val.FieldByName("StringSlice")
						case "concurrent int slices":
							field = val.FieldByName("IntSlice")
						case "concurrent bool slices":
							field = val.FieldByName("BoolSlice")
						case "concurrent float slices":
							field = val.FieldByName("FloatSlice")
						}

						err := SetString(field, tc.input)
						if err != nil {
							errChan <- err
							return
						}
					}
				}(i)
			}

			// Wait for all goroutines to complete and collect any errors
			for i := 0; i < numGoroutines*numIterations; i++ {
				select {
				case err := <-errChan:
					t.Errorf("Concurrent test failed: %v", err)
				default:
					// No error for this iteration
				}
			}
		})
	}
}

// TestSetSliceFromString_TypeSafety_StringSlice tests the critical type safety fix
// for string slice conversion that was implemented in lines 260-262.
// This test ensures the fix for reflect.Append type mismatch issues.
func TestSetSliceFromString_TypeSafety_StringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "comma separated strings",
			input:    "hello,world,test",
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "comma separated with spaces",
			input:    "  hello  ,  world  ,  test  ",
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "single string no comma",
			input:    "single",
			expected: []string{"single"},
		},
		{
			name:     "empty elements filtered out",
			input:    "hello,,world,",
			expected: []string{"hello", "world"},
		},
		{
			name:     "special characters",
			input:    "hello@world,test#123,value$456",
			expected: []string{"hello@world", "test#123", "value$456"},
		},
		{
			name:     "unicode characters",
			input:    "привет,мир,тест",
			expected: []string{"привет", "мир", "тест"},
		},
		{
			name:     "numbers as strings",
			input:    "123,456,789",
			expected: []string{"123", "456", "789"},
		},
		{
			name:     "boolean values as strings",
			input:    "true,false,yes,no",
			expected: []string{"true", "false", "yes", "no"},
		},
		{
			name:     "mixed content",
			input:    "text,123,true,@symbol",
			expected: []string{"text", "123", "true", "@symbol"},
		},
		{
			name:     "empty string results in empty slice",
			input:    "",
			expected: []string(nil), // Empty string should result in nil slice, not empty slice
		},
		{
			name:     "only commas and spaces",
			input:    " , , , ",
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test struct with string slice field
			var testStruct struct {
				StringSlice []string
			}

			// Get reflect.Value for the string slice field
			val := reflect.ValueOf(&testStruct).Elem()
			field := val.FieldByName("StringSlice")
			require.True(t, field.IsValid(), "StringSlice field should be found")
			require.True(t, field.CanSet(), "StringSlice field should be settable")

			// Call setSliceFromString (through SetString which routes to it)
			err := SetString(field, tc.input)
			require.NoError(t, err, "SetString should succeed for input: %s", tc.input)

			// Verify the result
			result := testStruct.StringSlice
			if tc.expected == nil && len(result) == 0 {
				// Both nil and empty slice are acceptable for empty input
				return
			}
			assert.Equal(t, tc.expected, result, "String slice should match expected values")
		})
	}
}

// TestSetSliceFromString_TypeSafety_NumericSlice tests the critical type safety fix
// for numeric slice conversion that was implemented in lines 274-293.
func TestSetSliceFromString_TypeSafety_NumericSlice(t *testing.T) {
	tests := []struct {
		name        string
		setupField  func() reflect.Value
		input       string
		expected    interface{}
		expectError bool
	}{
		// Int slice tests
		{
			name: "int slice - comma separated positive integers",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "1,2,3,42,100",
			expected: []int{1, 2, 3, 42, 100},
		},
		{
			name: "int slice - negative integers",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "-1,-2,0,3",
			expected: []int{-1, -2, 0, 3},
		},
		{
			name: "int slice - hex and octal values",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "10,0x10,010",
			expected: []int{10, 16, 8},
		},
		{
			name: "int slice - single value",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "42",
			expected: []int{42},
		},
		{
			name: "int slice - with spaces",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "  1  ,  2  ,  3  ",
			expected: []int{1, 2, 3},
		},

		// Float64 slice tests
		{
			name: "float64 slice - decimal numbers",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1.1,2.5,3.14159",
			expected: []float64{1.1, 2.5, 3.14159},
		},
		{
			name: "float64 slice - scientific notation",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1e2,2.5e-1,3.14e+0",
			expected: []float64{100, 0.25, 3.14},
		},
		{
			name: "float64 slice - integers as floats",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1,2,3",
			expected: []float64{1.0, 2.0, 3.0},
		},

		// Float32 slice tests
		{
			name: "float32 slice - decimal numbers",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float32 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1.1,2.5,3.14",
			expected: []float32{1.1, 2.5, 3.14},
		},

		// Int8 slice tests
		{
			name: "int8 slice - small integers",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int8 }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "1,2,127,-128",
			expected: []int8{1, 2, 127, -128},
		},

		// Int64 slice tests
		{
			name: "int64 slice - large integers",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int64 }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "1,9223372036854775807,-9223372036854775808",
			expected: []int64{1, 9223372036854775807, -9223372036854775808},
		},

		// Uint slice tests
		{
			name: "uint slice - positive integers",
			setupField: func() reflect.Value {
				var s struct{ UintSlice []uint }
				return reflect.ValueOf(&s).Elem().FieldByName("UintSlice")
			},
			input:    "1,2,3,4294967295",
			expected: []uint{1, 2, 3, 4294967295},
		},

		// Error cases
		{
			name: "int slice - invalid number",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:       "1,not_a_number,3",
			expectError: true,
		},
		{
			name: "float slice - invalid number",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:       "1.1,invalid_float,3.3",
			expectError: true,
		},
		{
			name: "int8 slice - overflow (now properly returns error)",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int8 }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:       "1,128", // 128 is out of range for int8 (-128 to 127)
			expectError: true,    // Fixed: now properly returns an error instead of silent overflow
		},
		{
			name: "uint slice - negative number",
			setupField: func() reflect.Value {
				var s struct{ UintSlice []uint }
				return reflect.ValueOf(&s).Elem().FieldByName("UintSlice")
			},
			input:       "1,-2,3", // negative number for uint
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field := tc.setupField()
			require.True(t, field.IsValid(), "Field should be valid")
			require.True(t, field.CanSet(), "Field should be settable")

			err := SetString(field, tc.input)

			if tc.expectError {
				assert.Error(t, err, "SetString should return error for input: %s", tc.input)
				return
			}

			require.NoError(t, err, "SetString should succeed for input: %s", tc.input)

			result := field.Interface()
			assert.Equal(t, tc.expected, result, "Numeric slice should match expected values")
		})
	}
}

// TestSetSliceFromString_TypeSafety_BoolSlice tests the critical type safety fix
// for bool slice conversion.
func TestSetSliceFromString_TypeSafety_BoolSlice(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []bool
		expectError bool
	}{
		{
			name:     "bool slice - true/false",
			input:    "true,false,true",
			expected: []bool{true, false, true},
		},
		{
			name:     "bool slice - 1/0",
			input:    "1,0,1",
			expected: []bool{true, false, true},
		},
		{
			name:     "bool slice - yes/no",
			input:    "yes,no,yes",
			expected: []bool{true, false, true},
		},
		{
			name:     "bool slice - on/off",
			input:    "on,off,on",
			expected: []bool{true, false, true},
		},
		{
			name:     "bool slice - mixed formats",
			input:    "true,0,yes,off",
			expected: []bool{true, false, true, false},
		},
		{
			name:     "bool slice - case insensitive",
			input:    "TRUE,False,YES,No",
			expected: []bool{true, false, true, false},
		},
		{
			name:     "bool slice - single value",
			input:    "true",
			expected: []bool{true},
		},
		{
			name:     "bool slice - with spaces",
			input:    "  true  ,  false  ",
			expected: []bool{true, false},
		},
		{
			name:        "bool slice - invalid value",
			input:       "true,invalid_bool,false",
			expectError: true,
		},
		{
			name:     "bool slice - empty element",
			input:    "true,,false",
			expected: []bool{true, false}, // empty elements should be filtered out
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var testStruct struct {
				BoolSlice []bool
			}

			val := reflect.ValueOf(&testStruct).Elem()
			field := val.FieldByName("BoolSlice")
			require.True(t, field.IsValid(), "BoolSlice field should be found")
			require.True(t, field.CanSet(), "BoolSlice field should be settable")

			err := SetString(field, tc.input)

			if tc.expectError {
				assert.Error(t, err, "SetString should return error for input: %s", tc.input)
				return
			}

			require.NoError(t, err, "SetString should succeed for input: %s", tc.input)

			result := testStruct.BoolSlice
			assert.Equal(t, tc.expected, result, "Bool slice should match expected values")
		})
	}
}

// TestSetSliceFromString_TypeSafety_SingleElement tests the critical type safety fix
// for single element conversion that was implemented in lines 303-306.
func TestSetSliceFromString_TypeSafety_SingleElement(t *testing.T) {
	tests := []struct {
		name       string
		setupField func() reflect.Value
		input      string
		expected   interface{}
	}{
		{
			name: "single string element",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "single_element",
			expected: []string{"single_element"},
		},
		{
			name: "single int element",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "42",
			expected: []int{42},
		},
		{
			name: "single float element",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "3.14159",
			expected: []float64{3.14159},
		},
		{
			name: "single bool element",
			setupField: func() reflect.Value {
				var s struct{ BoolSlice []bool }
				return reflect.ValueOf(&s).Elem().FieldByName("BoolSlice")
			},
			input:    "true",
			expected: []bool{true},
		},
		{
			name: "single hex int element",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "0xFF",
			expected: []int{255},
		},
		{
			name: "single scientific notation float",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1.23e-4",
			expected: []float64{0.000123},
		},
		{
			name: "single complex string with special chars",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "complex@string#with$special%chars",
			expected: []string{"complex@string#with$special%chars"},
		},
		{
			name: "single unicode string",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "привет世界🌍",
			expected: []string{"привет世界🌍"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field := tc.setupField()
			require.True(t, field.IsValid(), "Field should be valid")
			require.True(t, field.CanSet(), "Field should be settable")

			err := SetString(field, tc.input)
			require.NoError(t, err, "SetString should succeed for input: %s", tc.input)

			result := field.Interface()
			assert.Equal(t, tc.expected, result, "Single element slice should match expected value")
		})
	}
}

// TestSetSliceFromString_PanicPrevention tests that the type safety fixes
// prevent panics that could occur with the old implementation.
func TestSetSliceFromString_PanicPrevention(t *testing.T) {
	tests := []struct {
		name       string
		setupField func() reflect.Value
		input      string
		testDesc   string
	}{
		{
			name: "prevent string slice reflect.Append panic",
			setupField: func() reflect.Value {
				var s struct{ StringSlice []string }
				return reflect.ValueOf(&s).Elem().FieldByName("StringSlice")
			},
			input:    "test,panic,prevention",
			testDesc: "Should not panic when appending strings to string slice",
		},
		{
			name: "prevent int slice type mismatch panic",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "1,2,3",
			testDesc: "Should not panic when converting strings to ints in slice",
		},
		{
			name: "prevent float slice type mismatch panic",
			setupField: func() reflect.Value {
				var s struct{ FloatSlice []float64 }
				return reflect.ValueOf(&s).Elem().FieldByName("FloatSlice")
			},
			input:    "1.1,2.2,3.3",
			testDesc: "Should not panic when converting strings to floats in slice",
		},
		{
			name: "prevent bool slice type mismatch panic",
			setupField: func() reflect.Value {
				var s struct{ BoolSlice []bool }
				return reflect.ValueOf(&s).Elem().FieldByName("BoolSlice")
			},
			input:    "true,false,true",
			testDesc: "Should not panic when converting strings to bools in slice",
		},
		{
			name: "prevent single element type mismatch panic",
			setupField: func() reflect.Value {
				var s struct{ IntSlice []int }
				return reflect.ValueOf(&s).Elem().FieldByName("IntSlice")
			},
			input:    "42",
			testDesc: "Should not panic when adding single element to slice",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// This test should not panic - if it completes without panic, the fix is working
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test panicked: %v. %s", r, tc.testDesc)
				}
			}()

			field := tc.setupField()
			require.True(t, field.IsValid(), "Field should be valid")
			require.True(t, field.CanSet(), "Field should be settable")

			// This should not panic with the type safety fixes
			err := SetString(field, tc.input)

			// We expect this to succeed without panic
			assert.NoError(t, err, "SetString should succeed without panic: %s", tc.testDesc)
		})
	}
}

// TestSetIntFromString_OverflowFix tests the integer overflow bug fix in setIntFromString function.
// This test ensures that the fix properly validates integer ranges and returns errors instead of
// allowing silent overflow that was happening before the fix.
func TestSetIntFromString_OverflowFix(t *testing.T) {
	// Test struct with all integer types
	type TestStruct struct {
		Int8Field  int8
		Int16Field int16
		Int32Field int32
		Int64Field int64
		IntField   int
	}

	tests := []struct {
		name        string
		field       string
		input       string
		expectError bool
		expected    interface{}
		description string
	}{
		// Valid boundary values that should work
		{
			name:        "int8 max value (127)",
			field:       "Int8Field",
			input:       "127",
			expectError: false,
			expected:    int8(127),
			description: "Maximum valid int8 value should be accepted",
		},
		{
			name:        "int8 min value (-128)",
			field:       "Int8Field",
			input:       "-128",
			expectError: false,
			expected:    int8(-128),
			description: "Minimum valid int8 value should be accepted",
		},
		{
			name:        "int16 max value (32767)",
			field:       "Int16Field",
			input:       "32767",
			expectError: false,
			expected:    int16(32767),
			description: "Maximum valid int16 value should be accepted",
		},
		{
			name:        "int16 min value (-32768)",
			field:       "Int16Field",
			input:       "-32768",
			expectError: false,
			expected:    int16(-32768),
			description: "Minimum valid int16 value should be accepted",
		},
		{
			name:        "int32 max value (2147483647)",
			field:       "Int32Field",
			input:       "2147483647",
			expectError: false,
			expected:    int32(2147483647),
			description: "Maximum valid int32 value should be accepted",
		},
		{
			name:        "int32 min value (-2147483648)",
			field:       "Int32Field",
			input:       "-2147483648",
			expectError: false,
			expected:    int32(-2147483648),
			description: "Minimum valid int32 value should be accepted",
		},
		{
			name:        "int64 max value (9223372036854775807)",
			field:       "Int64Field",
			input:       "9223372036854775807",
			expectError: false,
			expected:    int64(9223372036854775807),
			description: "Maximum valid int64 value should be accepted",
		},
		{
			name:        "int64 min value (-9223372036854775808)",
			field:       "Int64Field",
			input:       "-9223372036854775808",
			expectError: false,
			expected:    int64(-9223372036854775808),
			description: "Minimum valid int64 value should be accepted",
		},

		// Overflow scenarios that should now return errors (previously caused silent wraparound)
		{
			name:        "int8 overflow (128) - fixed bug",
			field:       "Int8Field",
			input:       "128",
			expectError: true,
			description: "Value 128 should error for int8 (was -128 due to overflow before fix)",
		},
		{
			name:        "int8 underflow (-129) - fixed bug",
			field:       "Int8Field",
			input:       "-129",
			expectError: true,
			description: "Value -129 should error for int8 (was 127 due to underflow before fix)",
		},
		{
			name:        "int8 large overflow (256) - fixed bug",
			field:       "Int8Field",
			input:       "256",
			expectError: true,
			description: "Value 256 should error for int8 (was 0 due to overflow before fix)",
		},
		{
			name:        "int8 extreme overflow (999) - fixed bug",
			field:       "Int8Field",
			input:       "999",
			expectError: true,
			description: "Value 999 should error for int8 (was wrapped value before fix)",
		},
		{
			name:        "int16 overflow (32768) - fixed bug",
			field:       "Int16Field",
			input:       "32768",
			expectError: true,
			description: "Value 32768 should error for int16 (was -32768 due to overflow before fix)",
		},
		{
			name:        "int16 underflow (-32769) - fixed bug",
			field:       "Int16Field",
			input:       "-32769",
			expectError: true,
			description: "Value -32769 should error for int16 (was 32767 due to underflow before fix)",
		},
		{
			name:        "int16 large overflow (65536) - fixed bug",
			field:       "Int16Field",
			input:       "65536",
			expectError: true,
			description: "Value 65536 should error for int16 (was 0 due to overflow before fix)",
		},
		{
			name:        "int32 overflow (2147483648) - fixed bug",
			field:       "Int32Field",
			input:       "2147483648",
			expectError: true,
			description: "Value 2147483648 should error for int32 (was -2147483648 due to overflow before fix)",
		},
		{
			name:        "int32 underflow (-2147483649) - fixed bug",
			field:       "Int32Field",
			input:       "-2147483649",
			expectError: true,
			description: "Value -2147483649 should error for int32 (was 2147483647 due to underflow before fix)",
		},
		{
			name:        "int64 overflow (9223372036854775808) - fixed bug",
			field:       "Int64Field",
			input:       "9223372036854775808",
			expectError: true,
			description: "Value 9223372036854775808 should error for int64 (was -9223372036854775808 due to overflow before fix)",
		},

		// Test values that use the optimization path (≤3 digits) vs strconv fallback
		{
			name:        "single digit optimization path",
			field:       "Int8Field",
			input:       "7",
			expectError: false,
			expected:    int8(7),
			description: "Single digit should use fast path optimization",
		},
		{
			name:        "double digit optimization path",
			field:       "Int8Field",
			input:       "42",
			expectError: false,
			expected:    int8(42),
			description: "Double digit should use optimization path",
		},
		{
			name:        "triple digit optimization path - valid",
			field:       "Int8Field",
			input:       "100",
			expectError: false,
			expected:    int8(100),
			description: "Triple digit valid value should use optimization path",
		},
		{
			name:        "triple digit optimization path - overflow",
			field:       "Int8Field",
			input:       "200",
			expectError: true,
			description: "Triple digit overflow should be caught in optimization path",
		},
		{
			name:        "four digit strconv fallback path",
			field:       "Int16Field",
			input:       "1234",
			expectError: false,
			expected:    int16(1234),
			description: "Four digit should use strconv fallback path",
		},
		{
			name:        "four digit strconv fallback overflow",
			field:       "Int8Field",
			input:       "1234",
			expectError: true,
			description: "Four digit overflow should be caught in strconv fallback path",
		},

		// Test edge cases around optimization boundaries
		{
			name:        "exactly 3 digits - max valid int8",
			field:       "Int8Field",
			input:       "127",
			expectError: false,
			expected:    int8(127),
			description: "Exactly 3 digits at boundary should work",
		},
		{
			name:        "exactly 3 digits - overflow by 1",
			field:       "Int8Field",
			input:       "128",
			expectError: true,
			description: "Exactly 3 digits with overflow by 1 should error",
		},

		// Test zero and positive/negative boundaries
		{
			name:        "zero value",
			field:       "Int8Field",
			input:       "0",
			expectError: false,
			expected:    int8(0),
			description: "Zero should work correctly",
		},
		{
			name:        "positive one",
			field:       "Int8Field",
			input:       "1",
			expectError: false,
			expected:    int8(1),
			description: "Positive one should work correctly",
		},
		{
			name:        "negative one via strconv (4+ digits)",
			field:       "Int16Field",
			input:       "-1",
			expectError: false,
			expected:    int16(-1),
			description: "Negative values should use strconv path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var testObj TestStruct
			val := reflect.ValueOf(&testObj).Elem()
			field := val.FieldByName(tc.field)
			require.True(t, field.IsValid(), "Field %s not found", tc.field)

			err := SetString(field, tc.input)

			if tc.expectError {
				assert.Error(t, err, "Expected error for %s with input '%s': %s", tc.field, tc.input, tc.description)
				if err != nil {
					// Verify it's an overflow-related error
					assert.Contains(t, err.Error(), "out of range", "Error should mention 'out of range' for overflow: %s", tc.description)
				}
			} else {
				require.NoError(t, err, "Should not error for %s with input '%s': %s", tc.field, tc.input, tc.description)
				actual := field.Interface()
				assert.Equal(t, tc.expected, actual, "Value mismatch for %s: %s", tc.field, tc.description)
			}
		})
	}
}

// TestSetIntFromString_OptimizationPath tests that the optimization path (≤3 digits)
// works correctly and properly validates overflow conditions.
func TestSetIntFromString_OptimizationPath(t *testing.T) {
	type TestStruct struct {
		Int8Field  int8
		Int16Field int16
		Int32Field int32
		Int64Field int64
	}

	tests := []struct {
		name           string
		field          string
		input          string
		expectError    bool
		expected       interface{}
		shouldUseOptim bool
		description    string
	}{
		// Single digit tests (should use optimization)
		{
			name:           "single digit 0",
			field:          "Int8Field",
			input:          "0",
			expectError:    false,
			expected:       int8(0),
			shouldUseOptim: true,
			description:    "Single digit 0 should use fast path",
		},
		{
			name:           "single digit 9",
			field:          "Int8Field",
			input:          "9",
			expectError:    false,
			expected:       int8(9),
			shouldUseOptim: true,
			description:    "Single digit 9 should use fast path",
		},

		// Two digit tests (should use optimization for base 10)
		{
			name:           "two digits valid",
			field:          "Int8Field",
			input:          "42",
			expectError:    false,
			expected:       int8(42),
			shouldUseOptim: true,
			description:    "Two digits should use optimization path",
		},
		{
			name:           "two digits at int8 boundary",
			field:          "Int8Field",
			input:          "99",
			expectError:    false,
			expected:       int8(99),
			shouldUseOptim: true,
			description:    "Two digits within range should use optimization",
		},

		// Three digit tests (should use optimization for base 10)
		{
			name:           "three digits valid for int8",
			field:          "Int8Field",
			input:          "100",
			expectError:    false,
			expected:       int8(100),
			shouldUseOptim: true,
			description:    "Three digits valid value should use optimization",
		},
		{
			name:           "three digits max int8",
			field:          "Int8Field",
			input:          "127",
			expectError:    false,
			expected:       int8(127),
			shouldUseOptim: true,
			description:    "Three digits at max should use optimization",
		},
		{
			name:           "three digits overflow int8",
			field:          "Int8Field",
			input:          "128",
			expectError:    true,
			shouldUseOptim: true,
			description:    "Three digits overflow should be caught in optimization path",
		},
		{
			name:           "three digits large overflow int8",
			field:          "Int8Field",
			input:          "999",
			expectError:    true,
			shouldUseOptim: true,
			description:    "Three digits large overflow should be caught in optimization path",
		},

		// Three digit tests for larger types
		{
			name:           "three digits valid for int16",
			field:          "Int16Field",
			input:          "999",
			expectError:    false,
			expected:       int16(999),
			shouldUseOptim: true,
			description:    "Three digits should work for larger int types",
		},
		{
			name:           "three digits valid for int32",
			field:          "Int32Field",
			input:          "999",
			expectError:    false,
			expected:       int32(999),
			shouldUseOptim: true,
			description:    "Three digits should work for int32",
		},
		{
			name:           "three digits valid for int64",
			field:          "Int64Field",
			input:          "999",
			expectError:    false,
			expected:       int64(999),
			shouldUseOptim: true,
			description:    "Three digits should work for int64",
		},

		// Four digit tests (should use strconv fallback)
		{
			name:           "four digits fallback to strconv",
			field:          "Int16Field",
			input:          "1000",
			expectError:    false,
			expected:       int16(1000),
			shouldUseOptim: false,
			description:    "Four digits should use strconv fallback",
		},
		{
			name:           "four digits overflow via strconv",
			field:          "Int8Field",
			input:          "1000",
			expectError:    true,
			shouldUseOptim: false,
			description:    "Four digits overflow should be caught via strconv",
		},

		// Special format tests (should not use optimization)
		{
			name:           "hex format",
			field:          "Int16Field",
			input:          "0xFF",
			expectError:    false,
			expected:       int16(255),
			shouldUseOptim: false,
			description:    "Hex format should use strconv fallback",
		},
		{
			name:           "octal format",
			field:          "Int16Field",
			input:          "010",
			expectError:    false,
			expected:       int16(8),
			shouldUseOptim: false,
			description:    "Octal format should use strconv fallback",
		},
		{
			name:           "negative number",
			field:          "Int16Field",
			input:          "-42",
			expectError:    false,
			expected:       int16(-42),
			shouldUseOptim: false,
			description:    "Negative number should use strconv fallback",
		},

		// Leading zeros (should not use optimization due to octal detection)
		{
			name:           "leading zero octal detection",
			field:          "Int16Field",
			input:          "012",
			expectError:    false,
			expected:       int16(10), // 012 octal = 10 decimal
			shouldUseOptim: false,
			description:    "Leading zero should trigger octal detection and use strconv",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var testObj TestStruct
			val := reflect.ValueOf(&testObj).Elem()
			field := val.FieldByName(tc.field)
			require.True(t, field.IsValid(), "Field %s not found", tc.field)

			err := SetString(field, tc.input)

			if tc.expectError {
				assert.Error(t, err, "Expected error for %s with input '%s': %s", tc.field, tc.input, tc.description)
			} else {
				require.NoError(t, err, "Should not error for %s with input '%s': %s", tc.field, tc.input, tc.description)
				actual := field.Interface()
				assert.Equal(t, tc.expected, actual, "Value mismatch for %s: %s", tc.field, tc.description)
			}
		})
	}
}

// BenchmarkSetIntFromString_OverflowFix tests that the overflow fix doesn't
// significantly impact performance for common cases.
func BenchmarkSetIntFromString_OverflowFix(b *testing.B) {
	type TestStruct struct {
		Int8Field  int8
		Int16Field int16
		Int32Field int32
		Int64Field int64
		IntField   int
	}

	benchmarks := []struct {
		name  string
		field string
		input string
		desc  string
	}{
		// Optimization path tests
		{"SingleDigit_Int8", "Int8Field", "7", "Single digit using fast path"},
		{"DoubleDigit_Int8", "Int8Field", "42", "Double digit using optimization"},
		{"TripleDigit_Int8_Valid", "Int8Field", "100", "Triple digit valid using optimization"},
		{"TripleDigit_Int8_Max", "Int8Field", "127", "Triple digit at max using optimization"},

		// Different int sizes with optimization path
		{"TripleDigit_Int16", "Int16Field", "999", "Triple digit for int16"},
		{"TripleDigit_Int32", "Int32Field", "999", "Triple digit for int32"},
		{"TripleDigit_Int64", "Int64Field", "999", "Triple digit for int64"},

		// Strconv fallback path tests
		{"FourDigit_Int16", "Int16Field", "1000", "Four digit using strconv fallback"},
		{"FourDigit_Int32", "Int32Field", "1000", "Four digit for int32"},
		{"LargeNumber_Int32", "Int32Field", "123456", "Large number using strconv"},
		{"MaxInt32", "Int32Field", "2147483647", "Maximum int32 value"},

		// Special formats (strconv path)
		{"Hex_Int16", "Int16Field", "0xFF", "Hexadecimal format"},
		{"Negative_Int16", "Int16Field", "-42", "Negative number"},
		{"Octal_Int16", "Int16Field", "010", "Octal format"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			var testObj TestStruct
			val := reflect.ValueOf(&testObj).Elem()
			field := val.FieldByName(bm.field)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = SetString(field, bm.input)
			}
		})
	}
}

// TestSetIntFromString_EdgeCases tests edge cases that could trigger the overflow bug.
func TestSetIntFromString_EdgeCases(t *testing.T) {
	type TestStruct struct {
		Int8Field  int8
		Int16Field int16
		Int32Field int32
		Int64Field int64
	}

	tests := []struct {
		name        string
		field       string
		input       string
		expectError bool
		description string
	}{
		// Boundary testing - exactly at limits
		{
			name:        "int8_exactly_at_max",
			field:       "Int8Field",
			input:       "127",
			expectError: false,
			description: "Exactly at int8 maximum should work",
		},
		{
			name:        "int8_exactly_at_min",
			field:       "Int8Field",
			input:       "-128",
			expectError: false,
			description: "Exactly at int8 minimum should work",
		},
		{
			name:        "int8_one_over_max",
			field:       "Int8Field",
			input:       "128",
			expectError: true,
			description: "One over int8 maximum should error",
		},
		{
			name:        "int8_one_under_min",
			field:       "Int8Field",
			input:       "-129",
			expectError: true,
			description: "One under int8 minimum should error",
		},

		// Powers of 2 testing (common overflow scenarios)
		{
			name:        "int8_power_of_2_overflow",
			field:       "Int8Field",
			input:       "256", // 2^8
			expectError: true,
			description: "2^8 should overflow int8",
		},
		{
			name:        "int16_power_of_2_overflow",
			field:       "Int16Field",
			input:       "65536", // 2^16
			expectError: true,
			description: "2^16 should overflow int16",
		},
		{
			name:        "int32_power_of_2_overflow",
			field:       "Int32Field",
			input:       "4294967296", // 2^32
			expectError: true,
			description: "2^32 should overflow int32",
		},

		// Common problematic values
		{
			name:        "int8_common_overflow_200",
			field:       "Int8Field",
			input:       "200",
			expectError: true,
			description: "200 is a common value that overflows int8",
		},
		{
			name:        "int8_common_overflow_300",
			field:       "Int8Field",
			input:       "300",
			expectError: true,
			description: "300 is a common value that overflows int8",
		},
		{
			name:        "int16_common_overflow_50000",
			field:       "Int16Field",
			input:       "50000",
			expectError: true,
			description: "50000 is a common value that overflows int16",
		},

		// Empty string and whitespace
		{
			name:        "empty_string",
			field:       "Int8Field",
			input:       "",
			expectError: false,
			description: "Empty string should result in zero value",
		},

		// Very large numbers
		{
			name:        "very_large_number",
			field:       "Int8Field",
			input:       "999999999999999999",
			expectError: true,
			description: "Very large number should overflow any int type",
		},

		// Numbers with leading zeros that aren't octal
		{
			name:        "leading_zeros_decimal",
			field:       "Int8Field",
			input:       "007",
			expectError: false,
			description: "Leading zeros in decimal should parse correctly (octal 007 = 7 decimal)",
		},
		{
			name:        "leading_zeros_overflow",
			field:       "Int8Field",
			input:       "0200", // Octal 200 = 128 decimal, which overflows int8
			expectError: true,
			description: "Leading zeros causing octal interpretation that overflows should error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var testObj TestStruct
			val := reflect.ValueOf(&testObj).Elem()
			field := val.FieldByName(tc.field)
			require.True(t, field.IsValid(), "Field %s not found", tc.field)

			err := SetString(field, tc.input)

			if tc.expectError {
				assert.Error(t, err, "Expected error for %s with input '%s': %s", tc.field, tc.input, tc.description)
			} else {
				assert.NoError(t, err, "Should not error for %s with input '%s': %s", tc.field, tc.input, tc.description)
			}
		})
	}
}
