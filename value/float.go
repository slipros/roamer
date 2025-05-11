// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"math"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	rerr "github.com/slipros/roamer/err"
)

// SetFloat converts a floating-point value to the appropriate type for the target field
// and sets the field's value. This function handles conversion to various types,
// including strings, booleans, all numeric types, complex types, and interfaces.
//
// The function is generic and works with any floating-point type (float32, float64).
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - number: The floating-point value to convert and set.
//
// Returns:
//   - error: An error if the conversion or setting fails, or nil if successful.
//
// Example usage (internal to the package):
//
//	// Convert and set a float64 value to a string field
//	stringField := reflect.ValueOf(&myStruct).Elem().FieldByName("Score")
//	if err := SetFloat(stringField, 42.5); err != nil {
//	    return err
//	}
//
//	// Convert and set a float32 value to an int field
//	intField := reflect.ValueOf(&myStruct).Elem().FieldByName("RoundedScore")
//	if err := SetFloat(intField, float32(42.5)); err != nil {
//	    return err // Will set 42
//	}
func SetFloat[F constraints.Float](field reflect.Value, number F) error {
	// Check if the field is settable
	if !field.CanSet() {
		return errors.Errorf("field of type %s is not settable", field.Type())
	}

	// Check for special float values (NaN and Infinity)
	floatVal := float64(number)
	isSpecial := math.IsNaN(floatVal) || math.IsInf(floatVal, 0)

	switch field.Kind() {
	case reflect.String:
		// Format float based on value:
		// - Use normal decimal notation for most numbers
		// - Use scientific notation only for very large or small numbers
		var str string
		// Handle special cases explicitly
		switch {
		case math.IsNaN(floatVal):
			str = "NaN"
		case math.IsInf(floatVal, 1):
			str = "+Inf"
		case math.IsInf(floatVal, -1):
			str = "-Inf"
		default:
			abs := math.Abs(floatVal)
			if (abs >= 0.0001 && abs < 10000000) || floatVal == 0 {
				// Use regular decimal format for normal range numbers
				str = strconv.FormatFloat(floatVal, 'f', -1, 64)
			} else {
				// Use scientific notation for very large or small numbers
				str = strconv.FormatFloat(floatVal, 'E', -1, 64)
			}
		}
		field.SetString(str)
		return nil

	case reflect.Bool:
		// Convert float to bool (true if > 0, false otherwise)
		field.SetBool(number > 0)
		return nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		// Handle special cases for int types
		if isSpecial {
			if math.IsNaN(floatVal) {
				return errors.Errorf("cannot convert NaN to integer type %s", field.Type())
			}
			if math.IsInf(floatVal, 1) {
				return errors.Errorf("cannot convert +Infinity to integer type %s", field.Type())
			}
			if math.IsInf(floatVal, -1) {
				return errors.Errorf("cannot convert -Infinity to integer type %s", field.Type())
			}
		}

		// Check if value is too small and will be truncated to zero
		if math.Abs(floatVal) < 1.0 && floatVal != 0.0 {
			// This is not an error, but a warning that could be logged in a real-world app
			// For now, just proceed with truncation
		}

		// Check if the value is within int64 range
		if floatVal > float64(math.MaxInt64) || floatVal < float64(math.MinInt64) {
			return errors.Errorf("value %v is outside the range of any signed integer type", floatVal)
		}

		// Check for specific int sizes
		int64Val := int64(floatVal)
		if err := checkSignedIntegerRange(field, int64Val); err != nil {
			return err
		}

		field.SetInt(int64Val)
		return nil

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		// Handle special cases for uint types
		if isSpecial {
			if math.IsNaN(floatVal) {
				return errors.Errorf("cannot convert NaN to unsigned integer type %s", field.Type())
			}
			if math.IsInf(floatVal, 1) {
				return errors.Errorf("cannot convert +Infinity to unsigned integer type %s", field.Type())
			}
			if math.IsInf(floatVal, -1) {
				return errors.Errorf("cannot convert -Infinity to unsigned integer type %s", field.Type())
			}
		}

		// Check for negative float values
		if number < 0 {
			return errors.Errorf("cannot set negative value %v to unsigned type %s", number, field.Type())
		}

		// Check if value is too small and will be truncated to zero
		if floatVal > 0 && floatVal < 1.0 {
			// This is not an error, but a warning that could be logged in a real-world app
			// For now, just proceed with truncation
		}

		// Check if the value is within uint64 range
		if floatVal > float64(math.MaxUint64) {
			return errors.Errorf("value %v is outside the range of any unsigned integer type", floatVal)
		}

		// Convert to uint64 and check range for specific uint sizes
		uint64Val := uint64(floatVal)
		if err := checkUnsignedIntegerRange(field, uint64Val); err != nil {
			return err
		}

		field.SetUint(uint64Val)
		return nil

	case reflect.Float32:
		// Handle special float values separately
		if isSpecial {
			if math.IsNaN(floatVal) {
				field.SetFloat(math.NaN())
				return nil
			}

			if math.IsInf(floatVal, 1) {
				field.SetFloat(math.Inf(1))
				return nil
			}

			if math.IsInf(floatVal, -1) {
				field.SetFloat(math.Inf(-1))
				return nil
			}
		}

		// Check for potential float32 overflow
		if floatVal > math.MaxFloat32 || floatVal < -math.MaxFloat32 {
			return errors.Errorf("value %v is outside the range of float32", floatVal)
		}

		field.SetFloat(floatVal)
		return nil

	case reflect.Float64:
		// No need to check range for float64 as it can hold any float32 or float64 value
		field.SetFloat(floatVal)
		return nil

	case reflect.Complex64, reflect.Complex128:
		// Set float value to the real part, imaginary part is 0
		// Handle special values
		if isSpecial {
			var complexVal complex128
			switch {
			case math.IsNaN(floatVal):
				// For NaN, set both real and imaginary parts to NaN
				complexVal = complex(math.NaN(), math.NaN())
			case math.IsInf(floatVal, 1):
				complexVal = complex(math.Inf(1), 0)
			case math.IsInf(floatVal, -1):
				complexVal = complex(math.Inf(-1), 0)
			}
			field.SetComplex(complexVal)
			return nil
		}

		field.SetComplex(complex(floatVal, 0))
		return nil

	case reflect.Interface:
		// For interface{} fields, use the original float type (either float32 or float64)
		field.Set(reflect.ValueOf(number))
		return nil

	case reflect.Ptr:
		// For pointer fields, check if valid and then dereference and call recursively
		if field.IsNil() {
			// Initialize nil pointers
			field.Set(reflect.New(field.Type().Elem()))
		}
		return SetFloat[F](field.Elem(), number)
	}

	// If the field doesn't match any of the above types, return an error with more details
	return errors.Wrapf(rerr.NotSupported, "cannot set float value %v to field of type %s", number, field.Type())
}
