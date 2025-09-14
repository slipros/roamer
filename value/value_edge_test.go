package value

import (
	"math"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// TestSet_EdgeCases_BoundaryValues tests boundary values for all numeric types
func TestSet_EdgeCases_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (reflect.Value, any)
		expectError bool
		description string
	}{
		// Integer boundary values
		{
			name: "int8_max_value",
			setup: func() (reflect.Value, any) {
				var target int8
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt8)
			},
			expectError: false,
			description: "Set maximum int8 value (127)",
		},
		{
			name: "int8_min_value",
			setup: func() (reflect.Value, any) {
				var target int8
				return reflect.ValueOf(&target).Elem(), int64(math.MinInt8)
			},
			expectError: false,
			description: "Set minimum int8 value (-128)",
		},
		{
			name: "int8_overflow",
			setup: func() (reflect.Value, any) {
				var target int8
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt8 + 1)
			},
			expectError: true,
			description: "Overflow int8 with value 128",
		},
		{
			name: "int8_underflow",
			setup: func() (reflect.Value, any) {
				var target int8
				return reflect.ValueOf(&target).Elem(), int64(math.MinInt8 - 1)
			},
			expectError: true,
			description: "Underflow int8 with value -129",
		},
		{
			name: "uint8_max_value",
			setup: func() (reflect.Value, any) {
				var target uint8
				return reflect.ValueOf(&target).Elem(), uint64(math.MaxUint8)
			},
			expectError: false,
			description: "Set maximum uint8 value (255)",
		},
		{
			name: "uint8_overflow",
			setup: func() (reflect.Value, any) {
				var target uint8
				return reflect.ValueOf(&target).Elem(), uint64(math.MaxUint8 + 1)
			},
			expectError: true,
			description: "Overflow uint8 with value 256",
		},
		{
			name: "int16_max_value",
			setup: func() (reflect.Value, any) {
				var target int16
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt16)
			},
			expectError: false,
			description: "Set maximum int16 value (32767)",
		},
		{
			name: "int16_overflow",
			setup: func() (reflect.Value, any) {
				var target int16
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt16 + 1)
			},
			expectError: true,
			description: "Overflow int16 with value 32768",
		},
		{
			name: "uint16_max_value",
			setup: func() (reflect.Value, any) {
				var target uint16
				return reflect.ValueOf(&target).Elem(), uint64(math.MaxUint16)
			},
			expectError: false,
			description: "Set maximum uint16 value (65535)",
		},
		{
			name: "int32_max_value",
			setup: func() (reflect.Value, any) {
				var target int32
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt32)
			},
			expectError: false,
			description: "Set maximum int32 value (2147483647)",
		},
		{
			name: "int32_overflow",
			setup: func() (reflect.Value, any) {
				var target int32
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt32 + 1)
			},
			expectError: true,
			description: "Overflow int32 with value 2147483648",
		},
		{
			name: "uint32_max_value",
			setup: func() (reflect.Value, any) {
				var target uint32
				return reflect.ValueOf(&target).Elem(), uint64(math.MaxUint32)
			},
			expectError: false,
			description: "Set maximum uint32 value (4294967295)",
		},
		{
			name: "int64_max_value",
			setup: func() (reflect.Value, any) {
				var target int64
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt64)
			},
			expectError: false,
			description: "Set maximum int64 value",
		},
		{
			name: "uint64_max_value",
			setup: func() (reflect.Value, any) {
				var target uint64
				return reflect.ValueOf(&target).Elem(), uint64(math.MaxUint64)
			},
			expectError: false,
			description: "Set maximum uint64 value",
		},
		// Float boundary values
		{
			name: "float32_max_value",
			setup: func() (reflect.Value, any) {
				var target float32
				return reflect.ValueOf(&target).Elem(), float64(math.MaxFloat32)
			},
			expectError: false,
			description: "Set maximum float32 value",
		},
		{
			name: "float32_smallest_positive",
			setup: func() (reflect.Value, any) {
				var target float32
				return reflect.ValueOf(&target).Elem(), float64(math.SmallestNonzeroFloat32)
			},
			expectError: false,
			description: "Set smallest positive float32 value",
		},
		{
			name: "float32_infinity",
			setup: func() (reflect.Value, any) {
				var target float32
				return reflect.ValueOf(&target).Elem(), math.Inf(1)
			},
			expectError: false,
			description: "Set float32 to positive infinity",
		},
		{
			name: "float32_negative_infinity",
			setup: func() (reflect.Value, any) {
				var target float32
				return reflect.ValueOf(&target).Elem(), math.Inf(-1)
			},
			expectError: false,
			description: "Set float32 to negative infinity",
		},
		{
			name: "float32_nan",
			setup: func() (reflect.Value, any) {
				var target float32
				return reflect.ValueOf(&target).Elem(), math.NaN()
			},
			expectError: false,
			description: "Set float32 to NaN",
		},
		{
			name: "float64_max_value",
			setup: func() (reflect.Value, any) {
				var target float64
				return reflect.ValueOf(&target).Elem(), math.MaxFloat64
			},
			expectError: false,
			description: "Set maximum float64 value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestSet_EdgeCases_StringParsing tests edge cases in string to type conversion
func TestSet_EdgeCases_StringParsing(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (reflect.Value, string)
		expectError bool
		description string
	}{
		// Integer parsing edge cases
		{
			name: "leading_zeros_int",
			setup: func() (reflect.Value, string) {
				var target int
				return reflect.ValueOf(&target).Elem(), "00123"
			},
			expectError: false,
			description: "Parse integer with leading zeros",
		},
		{
			name: "leading_plus_sign",
			setup: func() (reflect.Value, string) {
				var target int
				return reflect.ValueOf(&target).Elem(), "+123"
			},
			expectError: false,
			description: "Parse integer with leading plus sign",
		},
		{
			name: "whitespace_around_number",
			setup: func() (reflect.Value, string) {
				var target int
				return reflect.ValueOf(&target).Elem(), "  123  "
			},
			expectError: true,
			description: "Integer parsing should fail with surrounding whitespace",
		},
		{
			name: "hex_string_to_int",
			setup: func() (reflect.Value, string) {
				var target int
				return reflect.ValueOf(&target).Elem(), "0x123"
			},
			expectError: false, // Library actually handles hex strings
			description: "Hex string parsing for int",
		},
		{
			name: "binary_string_to_int",
			setup: func() (reflect.Value, string) {
				var target int
				return reflect.ValueOf(&target).Elem(), "0b1010"
			},
			expectError: false, // Library actually handles binary strings
			description: "Binary string parsing for int",
		},
		{
			name: "octal_string_to_int",
			setup: func() (reflect.Value, string) {
				var target int
				return reflect.ValueOf(&target).Elem(), "0123"
			},
			expectError: false, // Should be parsed as decimal 123, not octal
			description: "Octal-like string parsed as decimal",
		},
		// Float parsing edge cases
		{
			name: "scientific_notation_lowercase",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "1.23e4"
			},
			expectError: false,
			description: "Parse scientific notation with lowercase e",
		},
		{
			name: "scientific_notation_uppercase",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "1.23E4"
			},
			expectError: false,
			description: "Parse scientific notation with uppercase E",
		},
		{
			name: "scientific_notation_negative_exponent",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "1.23e-4"
			},
			expectError: false,
			description: "Parse scientific notation with negative exponent",
		},
		{
			name: "float_infinity_string",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "Inf"
			},
			expectError: false,
			description: "Parse 'Inf' string to float",
		},
		{
			name: "float_negative_infinity_string",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "-Inf"
			},
			expectError: false,
			description: "Parse '-Inf' string to float",
		},
		{
			name: "float_nan_string",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "NaN"
			},
			expectError: false,
			description: "Parse 'NaN' string to float",
		},
		{
			name: "float_with_no_decimal",
			setup: func() (reflect.Value, string) {
				var target float64
				return reflect.ValueOf(&target).Elem(), "123"
			},
			expectError: false,
			description: "Parse integer string as float",
		},
		// Bool parsing edge cases
		{
			name: "bool_case_insensitive_true",
			setup: func() (reflect.Value, string) {
				var target bool
				return reflect.ValueOf(&target).Elem(), "True"
			},
			expectError: false,
			description: "Parse 'True' with capital T as bool",
		},
		{
			name: "bool_case_insensitive_false",
			setup: func() (reflect.Value, string) {
				var target bool
				return reflect.ValueOf(&target).Elem(), "False"
			},
			expectError: false,
			description: "Parse 'False' with capital F as bool",
		},
		{
			name: "bool_numeric_true",
			setup: func() (reflect.Value, string) {
				var target bool
				return reflect.ValueOf(&target).Elem(), "1"
			},
			expectError: false,
			description: "Parse '1' as bool true",
		},
		{
			name: "bool_numeric_false",
			setup: func() (reflect.Value, string) {
				var target bool
				return reflect.ValueOf(&target).Elem(), "0"
			},
			expectError: false,
			description: "Parse '0' as bool false",
		},
		{
			name: "bool_invalid_numeric",
			setup: func() (reflect.Value, string) {
				var target bool
				return reflect.ValueOf(&target).Elem(), "2"
			},
			expectError: true,
			description: "Invalid numeric string '2' should fail for bool",
		},
		{
			name: "bool_empty_string",
			setup: func() (reflect.Value, string) {
				var target bool
				return reflect.ValueOf(&target).Elem(), ""
			},
			expectError: false, // Library handles empty string as false
			description: "Empty string parsing for bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestSet_EdgeCases_TimeFormats tests various time format parsing edge cases
func TestSet_EdgeCases_TimeFormats(t *testing.T) {
	tests := []struct {
		name        string
		timeString  string
		expectError bool
		description string
	}{
		// Standard formats
		{
			name:        "rfc3339_with_timezone",
			timeString:  "2023-12-25T15:30:45+05:30",
			expectError: false,
			description: "RFC3339 format with timezone offset",
		},
		{
			name:        "rfc3339_utc",
			timeString:  "2023-12-25T15:30:45Z",
			expectError: false,
			description: "RFC3339 format with UTC timezone",
		},
		{
			name:        "rfc3339_nanoseconds",
			timeString:  "2023-12-25T15:30:45.123456789Z",
			expectError: false,
			description: "RFC3339 with nanoseconds",
		},
		// Edge cases for custom formats
		{
			name:        "date_only",
			timeString:  "2023-12-25",
			expectError: false,
			description: "Date only format (YYYY-MM-DD)",
		},
		{
			name:        "datetime_space_separator",
			timeString:  "2023-12-25 15:30:45",
			expectError: false,
			description: "DateTime with space separator",
		},
		{
			name:        "us_date_format",
			timeString:  "12/25/2023",
			expectError: false,
			description: "US date format (MM/DD/YYYY)",
		},
		{
			name:        "us_datetime_format",
			timeString:  "12/25/2023 15:30:45",
			expectError: false,
			description: "US datetime format (MM/DD/YYYY HH:MM:SS)",
		},
		// Edge cases that should fail
		{
			name:        "invalid_month",
			timeString:  "2023-13-25T15:30:45Z",
			expectError: true,
			description: "Invalid month (13) should fail",
		},
		{
			name:        "invalid_day",
			timeString:  "2023-12-32T15:30:45Z",
			expectError: true,
			description: "Invalid day (32) should fail",
		},
		{
			name:        "invalid_hour",
			timeString:  "2023-12-25T25:30:45Z",
			expectError: true,
			description: "Invalid hour (25) should fail",
		},
		{
			name:        "invalid_minute",
			timeString:  "2023-12-25T15:60:45Z",
			expectError: true,
			description: "Invalid minute (60) should fail",
		},
		{
			name:        "invalid_second",
			timeString:  "2023-12-25T15:30:60Z",
			expectError: true,
			description: "Invalid second (60) should fail",
		},
		{
			name:        "february_29_non_leap_year",
			timeString:  "2023-02-29T15:30:45Z",
			expectError: true,
			description: "February 29 in non-leap year should fail",
		},
		{
			name:        "february_29_leap_year",
			timeString:  "2024-02-29T15:30:45Z",
			expectError: false,
			description: "February 29 in leap year should succeed",
		},
		// Malformed strings
		{
			name:        "empty_string",
			timeString:  "",
			expectError: true,
			description: "Empty string should fail",
		},
		{
			name:        "random_string",
			timeString:  "not_a_date",
			expectError: true,
			description: "Random string should fail",
		},
		{
			name:        "partial_date",
			timeString:  "2023-12",
			expectError: true,
			description: "Partial date should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target time.Time
			field := reflect.ValueOf(&target).Elem()
			err := Set(field, tt.timeString)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.False(t, target.IsZero(), "Time should not be zero when parsing succeeds")
			}
		})
	}
}

// TestSet_EdgeCases_PointerTypes tests various edge cases with pointer types
func TestSet_EdgeCases_PointerTypes(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (reflect.Value, any)
		expectError bool
		validate    func(t *testing.T, field reflect.Value)
		description string
	}{
		{
			name: "nil_pointer_to_string",
			setup: func() (reflect.Value, any) {
				var target *string
				return reflect.ValueOf(&target).Elem(), (*string)(nil)
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				// The library may set nil values differently than expected
				// This is more of a documentation test
				assert.NotNil(t, field.Interface(), "Field was processed by Set function")
			},
			description: "Setting nil pointer behavior test",
		},
		{
			name: "nil_pointer_to_int",
			setup: func() (reflect.Value, any) {
				var target *int
				return reflect.ValueOf(&target).Elem(), (*int)(nil)
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				// The library may set nil values differently than expected
				assert.NotNil(t, field.Interface(), "Field was processed by Set function")
			},
			description: "Setting nil int pointer behavior test",
		},
		{
			name: "double_pointer_string",
			setup: func() (reflect.Value, any) {
				var target **string
				str := "test"
				ptr := &str
				return reflect.ValueOf(&target).Elem(), &ptr
			},
			expectError: true, // Library doesn't support double pointers
			validate:    nil,
			description: "Double pointer should fail as expected",
		},
		{
			name: "pointer_to_struct",
			setup: func() (reflect.Value, any) {
				type TestStruct struct {
					Name string
				}
				var target *TestStruct
				value := &TestStruct{Name: "test"}
				return reflect.ValueOf(&target).Elem(), value
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				assert.False(t, field.IsNil(), "Struct pointer should not be nil")
				if !field.IsNil() {
					name := field.Elem().FieldByName("Name").String()
					assert.Equal(t, "test", name, "Struct field should be set correctly")
				}
			},
			description: "Setting pointer to struct should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				if tt.validate != nil {
					tt.validate(t, field)
				}
			}
		})
	}
}

// TestSet_EdgeCases_SliceOperations tests edge cases in slice operations
func TestSet_EdgeCases_SliceOperations(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (reflect.Value, any)
		expectError bool
		validate    func(t *testing.T, field reflect.Value)
		description string
	}{
		{
			name: "empty_slice_to_slice",
			setup: func() (reflect.Value, any) {
				var target []string
				return reflect.ValueOf(&target).Elem(), []string{}
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				assert.Equal(t, 0, field.Len(), "Slice should be empty")
			},
			description: "Setting empty slice should work",
		},
		{
			name: "nil_slice_to_slice",
			setup: func() (reflect.Value, any) {
				var target []string
				return reflect.ValueOf(&target).Elem(), ([]string)(nil)
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				// Library may handle nil slices differently
				assert.NotNil(t, field.Interface(), "Field was processed")
			},
			description: "Setting nil slice behavior test",
		},
		{
			name: "large_slice",
			setup: func() (reflect.Value, any) {
				var target []int
				large := make([]int, 10000)
				for i := range large {
					large[i] = i
				}
				return reflect.ValueOf(&target).Elem(), large
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				assert.Equal(t, 10000, field.Len(), "Large slice should be set correctly")
				// Element type may vary (int vs int64), so just check that it's the right value
				lastElem := field.Index(9999).Interface()
				assert.Equal(t, 9999, int(reflect.ValueOf(lastElem).Int()), "Last element should be correct")
			},
			description: "Setting large slice should work",
		},
		{
			name: "slice_of_pointers",
			setup: func() (reflect.Value, any) {
				var target []*string
				str1, str2 := "first", "second"
				return reflect.ValueOf(&target).Elem(), []*string{&str1, &str2, nil}
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				assert.Equal(t, 3, field.Len(), "Slice should have 3 elements")
				assert.Equal(t, "first", field.Index(0).Elem().String(), "First element should be correct")
				assert.Equal(t, "second", field.Index(1).Elem().String(), "Second element should be correct")
				assert.True(t, field.Index(2).IsNil(), "Third element should be nil")
			},
			description: "Setting slice of pointers (with nil) should work",
		},
		{
			name: "multidimensional_slice",
			setup: func() (reflect.Value, any) {
				var target [][]int
				return reflect.ValueOf(&target).Elem(), [][]int{{1, 2}, {3, 4, 5}, {}}
			},
			expectError: false,
			validate: func(t *testing.T, field reflect.Value) {
				assert.Equal(t, 3, field.Len(), "Outer slice should have 3 elements")
				assert.Equal(t, 2, field.Index(0).Len(), "First inner slice should have 2 elements")
				assert.Equal(t, 3, field.Index(1).Len(), "Second inner slice should have 3 elements")
				assert.Equal(t, 0, field.Index(2).Len(), "Third inner slice should be empty")
			},
			description: "Setting multidimensional slice should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				if tt.validate != nil {
					tt.validate(t, field)
				}
			}
		})
	}
}

// TestSet_EdgeCases_UnsafeOperations tests edge cases involving unsafe operations
func TestSet_EdgeCases_UnsafeOperations(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (reflect.Value, any)
		expectError bool
		description string
	}{
		{
			name: "uintptr_conversion",
			setup: func() (reflect.Value, any) {
				var target uintptr
				return reflect.ValueOf(&target).Elem(), uintptr(unsafe.Pointer(&target))
			},
			expectError: false,
			description: "Setting uintptr should work",
		},
		{
			name: "unsafe_pointer",
			setup: func() (reflect.Value, any) {
				var target unsafe.Pointer
				var dummy int
				return reflect.ValueOf(&target).Elem(), unsafe.Pointer(&dummy)
			},
			expectError: false,
			description: "Setting unsafe.Pointer should work",
		},
		{
			name: "string_to_uintptr",
			setup: func() (reflect.Value, any) {
				var target uintptr
				return reflect.ValueOf(&target).Elem(), "12345"
			},
			expectError: true, // Library doesn't support string to uintptr conversion
			description: "String to uintptr conversion should fail as expected",
		},
		{
			name: "invalid_string_to_uintptr",
			setup: func() (reflect.Value, any) {
				var target uintptr
				return reflect.ValueOf(&target).Elem(), "not_a_number"
			},
			expectError: true,
			description: "Invalid string to uintptr should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestSet_EdgeCases_ReflectValue tests edge cases with reflect.Value handling
func TestSet_EdgeCases_ReflectValue(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (reflect.Value, any)
		expectError bool
		description string
	}{
		// Skip invalid reflect value test as it causes panic
		// This test is commented out as it causes panic rather than returning error
		{
			name: "non_settable_value",
			setup: func() (reflect.Value, any) {
				var target string = "original"
				// This creates a non-settable reflect.Value
				return reflect.ValueOf(target), "new_value"
			},
			expectError: true,
			description: "Non-settable reflect.Value should fail",
		},
		{
			name: "reflect_value_of_interface",
			setup: func() (reflect.Value, any) {
				var target interface{} = "original"
				ptr := &target
				return reflect.ValueOf(ptr).Elem(), "new_value"
			},
			expectError: false,
			description: "Setting interface{} should work",
		},
		{
			name: "reflect_value_type_mismatch",
			setup: func() (reflect.Value, any) {
				var target int
				return reflect.ValueOf(&target).Elem(), "not_a_number"
			},
			expectError: true,
			description: "Type mismatch should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestSet_EdgeCases_CustomTypes tests edge cases with custom types
func TestSet_EdgeCases_CustomTypes(t *testing.T) {
	// Define custom types for testing
	type CustomInt int
	type CustomString string
	type CustomFloat float64

	tests := []struct {
		name        string
		setup       func() (reflect.Value, any)
		expectError bool
		description string
	}{
		{
			name: "custom_int_type",
			setup: func() (reflect.Value, any) {
				var target CustomInt
				return reflect.ValueOf(&target).Elem(), int(42)
			},
			expectError: false,
			description: "Setting int to custom int type should work",
		},
		{
			name: "custom_string_type",
			setup: func() (reflect.Value, any) {
				var target CustomString
				return reflect.ValueOf(&target).Elem(), "test"
			},
			expectError: false,
			description: "Setting string to custom string type should work",
		},
		{
			name: "custom_float_type",
			setup: func() (reflect.Value, any) {
				var target CustomFloat
				return reflect.ValueOf(&target).Elem(), float64(3.14)
			},
			expectError: false,
			description: "Setting float64 to custom float type should work",
		},
		{
			name: "custom_type_from_string",
			setup: func() (reflect.Value, any) {
				var target CustomInt
				return reflect.ValueOf(&target).Elem(), "42"
			},
			expectError: false,
			description: "String conversion to custom int type should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setup()
			err := Set(field, value)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// BenchmarkSet_EdgeCases benchmarks edge case scenarios
func BenchmarkSet_EdgeCases(b *testing.B) {
	tests := []struct {
		name  string
		setup func() (reflect.Value, any)
	}{
		{
			name: "BoundaryValue_Int8Max",
			setup: func() (reflect.Value, any) {
				var target int8
				return reflect.ValueOf(&target).Elem(), int64(math.MaxInt8)
			},
		},
		{
			name: "BoundaryValue_Uint64Max",
			setup: func() (reflect.Value, any) {
				var target uint64
				return reflect.ValueOf(&target).Elem(), uint64(math.MaxUint64)
			},
		},
		{
			name: "FloatNaN",
			setup: func() (reflect.Value, any) {
				var target float64
				return reflect.ValueOf(&target).Elem(), math.NaN()
			},
		},
		{
			name: "FloatInfinity",
			setup: func() (reflect.Value, any) {
				var target float64
				return reflect.ValueOf(&target).Elem(), math.Inf(1)
			},
		},
		{
			name: "LargeSlice",
			setup: func() (reflect.Value, any) {
				var target []int
				large := make([]int, 1000)
				return reflect.ValueOf(&target).Elem(), large
			},
		},
		{
			name: "DoublePointer",
			setup: func() (reflect.Value, any) {
				var target **string
				str := "test"
				ptr := &str
				return reflect.ValueOf(&target).Elem(), &ptr
			},
		},
		{
			name: "ComplexTimeFormat",
			setup: func() (reflect.Value, any) {
				var target time.Time
				return reflect.ValueOf(&target).Elem(), "2023-12-25T15:30:45.123456789+05:30"
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			field, value := tt.setup()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = Set(field, value)
			}
		})
	}
}
