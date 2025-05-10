// Package value provides utilities for type conversion and setting values in Go structs.
// It is used by the roamer package to convert parsed values from HTTP requests
// into appropriate types for struct fields.
package value

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// Set assigns a value to a reflect.Value field, performing type conversion if necessary.
// This is a high-level function that handles various input types and target field types,
// dispatching to specialized setters based on the input type.
//
// The function handles:
// - Automatic initialization of nil pointers
// - Type conversion between compatible types
// - Special handling for strings, integers, floats, and string slices
// - Support for fmt.Stringer interface
//
// If the value cannot be set (e.g., incompatible types), an error is returned.
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - value: The value to set (can be of any type).
//
// Returns:
//   - error: An error if the value could not be set, or nil if successful.
func Set(field reflect.Value, value any) error {
	if field.Kind() == reflect.Pointer && field.IsNil() {
		// Initialize nil pointers
		field.Set(reflect.New(field.Type().Elem()))
		field = reflect.Indirect(field)
	}

	// Handle various types using specialized setters
	switch t := value.(type) {
	case string:
		return SetString(field, t)
	case *string:
		return SetString(field, *t)
	case int:
		return SetInteger(field, t)
	case *int:
		return SetInteger(field, *t)
	case int8:
		return SetInteger(field, t)
	case *int8:
		return SetInteger(field, *t)
	case int16:
		return SetInteger(field, t)
	case *int16:
		return SetInteger(field, *t)
	case int32:
		return SetInteger(field, t)
	case *int32:
		return SetInteger(field, *t)
	case int64:
		return SetInteger(field, t)
	case *int64:
		return SetInteger(field, *t)
	case uint:
		return SetInteger(field, t)
	case *uint:
		return SetInteger(field, *t)
	case uint8:
		return SetInteger(field, t)
	case *uint8:
		return SetInteger(field, *t)
	case uint16:
		return SetInteger(field, t)
	case *uint16:
		return SetInteger(field, *t)
	case uint32:
		return SetInteger(field, t)
	case *uint32:
		return SetInteger(field, *t)
	case uint64:
		return SetInteger(field, t)
	case *uint64:
		return SetInteger(field, *t)
	case float32:
		return SetFloat(field, t)
	case *float32:
		return SetFloat(field, *t)
	case float64:
		return SetFloat(field, t)
	case *float64:
		return SetFloat(field, *t)
	case []string:
		return SetSliceString(field, t)
	}

	// Handle general assignable types
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Pointer {
		// Dereference pointers
		valueType = valueType.Elem()
	}

	if field.Type().AssignableTo(valueType) {
		field.Set(reflect.Indirect(reflect.ValueOf(value)))
		return nil
	}

	// Handle types that implement fmt.Stringer
	if i, ok := value.(fmt.Stringer); ok {
		return SetString(field, i.String())
	}

	// If we reach here, the value couldn't be set
	return errors.WithStack(rerr.NotSupported)
}
