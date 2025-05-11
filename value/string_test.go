package value

import (
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
		StrField          string
		BoolField         bool
		IntField          int
		Int8Field         int8
		Int16Field        int16
		Int32Field        int32
		Int64Field        int64
		UintField         uint
		Uint8Field        uint8
		Uint16Field       uint16
		Uint32Field       uint32
		Uint64Field       uint64
		Float32Field      float32
		Float64Field      float64
		Complex64Field    complex64
		Complex128Field   complex128
		ByteSlice         []byte
		StringSlice       []string
		IntSlice          []int
		PtrField          *string
		TextUnmarshaler   UnmarshallerText
		BinaryUnmarshaler UnmarshallerBinary
		TimeField         time.Time
		InterfaceField    interface{}
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
		IntField           int
		UintField          uint
		FloatField         float64
		ComplexField       complex128
		TimeField          time.Time
		MapStringIntField  map[string]int
		MapIntStringField  map[int]string
		ChannelField       chan int
		FailingUnmarshaler failingUnmarshaler
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
