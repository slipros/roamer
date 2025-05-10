// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	rerr "github.com/slipros/roamer/err"
)

// SetFloat converts a floating-point value to the appropriate type for the target field
// and sets the field's value. This function handles conversion to various types,
// including strings, booleans, all numeric types, and interfaces.
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
	switch field.Kind() {
	case reflect.String:
		// Convert float to string with scientific notation
		field.SetString(strconv.FormatFloat(float64(number), 'E', -1, 64))
		return nil

	case reflect.Bool:
		// Convert float to bool (true if > 0, false otherwise)
		field.SetBool(number > 0)
		return nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		// Convert float to int (truncating any decimal part)
		field.SetInt(int64(number))
		return nil

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		// Convert float to uint (truncating any decimal part)
		field.SetUint(uint64(number))
		return nil

	case reflect.Float32, reflect.Float64:
		// Convert float to float (may involve precision loss for float32)
		field.SetFloat(float64(number))
		return nil

	case reflect.Interface:
		// For interface{} fields, just set the float value
		field.Set(reflect.ValueOf(number))
		return nil

	case reflect.Ptr:
		// For pointer fields, dereference and call recursively
		return SetFloat[F](field.Elem(), number)
	}

	// If the field doesn't match any of the above types, return an error
	return errors.WithStack(rerr.NotSupported)
}
