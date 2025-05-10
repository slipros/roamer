// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	rerr "github.com/slipros/roamer/err"
)

// SetInteger converts an integer value to the appropriate type for the target field
// and sets the field's value. This function handles conversion to various types,
// including strings, booleans, all numeric types, and interfaces.
//
// The function is generic and works with any integer type (int, int8, int16, int32, int64,
// uint, uint8, uint16, uint32, uint64).
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - number: The integer value to convert and set.
//
// Returns:
//   - error: An error if the conversion or setting fails, or nil if successful.
//
// Example usage (internal to the package):
//
//	// Convert and set an int value to a string field
//	stringField := reflect.ValueOf(&myStruct).Elem().FieldByName("Count")
//	if err := SetInteger(stringField, 42); err != nil {
//	    return err
//	}
//
//	// Convert and set a uint8 value to a float field
//	floatField := reflect.ValueOf(&myStruct).Elem().FieldByName("Score")
//	if err := SetInteger(floatField, uint8(100)); err != nil {
//	    return err
//	}
func SetInteger[I constraints.Integer](field reflect.Value, number I) error {
	switch field.Kind() {
	case reflect.String:
		// Convert integer to string
		field.SetString(strconv.Itoa(int(number)))
		return nil

	case reflect.Bool:
		// Convert integer to bool (true if > 0, false otherwise)
		field.SetBool(number > 0)
		return nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		// Convert integer to int (may involve truncation for larger types to smaller)
		field.SetInt(int64(number))
		return nil

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		// Convert integer to uint (may involve truncation for larger types to smaller)
		field.SetUint(uint64(number))
		return nil

	case reflect.Float32, reflect.Float64:
		// Convert integer to float
		field.SetFloat(float64(number))
		return nil

	case reflect.Interface:
		// For interface{} fields, just set the integer value
		field.Set(reflect.ValueOf(number))
		return nil

	case reflect.Ptr:
		// For pointer fields, dereference and call recursively
		return SetInteger[I](field.Elem(), number)
	}

	// If the field doesn't match any of the above types, return an error
	return errors.WithStack(rerr.NotSupported)
}
