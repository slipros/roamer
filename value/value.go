// Package value provides utilities for type conversion and setting values in Go structs.
// It is used by the roamer package to convert parsed values from HTTP requests
// into appropriate types for struct fields.
package value

import (
	"fmt"
	"reflect"
	"strconv"

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
// - Special handling for strings, integers, floats, booleans, and string slices
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
	// Check if field can be set
	if !field.CanSet() {
		return errors.Errorf("field of type %s is not settable", field.Type())
	}

	// Handle nil value early
	if value == nil {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	// Initialize and dereference nil pointers recursively
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			// Initialize the pointer with a new value of the appropriate type
			field.Set(reflect.New(field.Type().Elem()))
		}
		// Recursively call Set on the dereferenced pointer
		return Set(field.Elem(), value)
	}

	// Handle various types using specialized setters
	switch t := value.(type) {
	case string:
		return SetString(field, t)
	case *string:
		if t == nil {
			// For nil string pointers, set zero value
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetString(field, *t)
	case bool:
		// Add direct support for boolean values
		if field.Kind() == reflect.Bool {
			field.SetBool(t)
			return nil
		}
		return SetString(field, strconv.FormatBool(t))
	case *bool:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		if field.Kind() == reflect.Bool {
			field.SetBool(*t)
			return nil
		}
		return SetString(field, strconv.FormatBool(*t))
	case int:
		return SetInteger(field, t)
	case *int:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case int8:
		return SetInteger(field, t)
	case *int8:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case int16:
		return SetInteger(field, t)
	case *int16:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case int32:
		return SetInteger(field, t)
	case *int32:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case int64:
		return SetInteger(field, t)
	case *int64:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case uint:
		return SetInteger(field, t)
	case *uint:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case uint8:
		return SetInteger(field, t)
	case *uint8:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case uint16:
		return SetInteger(field, t)
	case *uint16:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case uint32:
		return SetInteger(field, t)
	case *uint32:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case uint64:
		return SetInteger(field, t)
	case *uint64:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetInteger(field, *t)
	case float32:
		return SetFloat(field, t)
	case *float32:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetFloat(field, *t)
	case float64:
		return SetFloat(field, t)
	case *float64:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return SetFloat(field, *t)
	case []string:
		return SetSliceString(field, t)
	case []any:
		// Handle []any differently based on the field type
		if field.Kind() == reflect.Slice {
			return handleInterfaceSlice(field, t)
		}
	}

	// Handle types that implement fmt.Stringer
	if stringer, ok := value.(fmt.Stringer); ok {
		return SetString(field, stringer.String())
	}

	// Handle general assignable types
	valueType := reflect.TypeOf(value)
	valueValue := reflect.ValueOf(value)

	// Handle pointers for general types
	if valueType.Kind() == reflect.Pointer {
		// Check if the pointer is nil
		if valueValue.IsNil() {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}

		// Dereference pointer and use indirect value
		valueType = valueType.Elem()
		valueValue = valueValue.Elem()
	}

	// Check if the target field's type can be assigned from the value's type
	if field.Type().AssignableTo(valueType) {
		field.Set(valueValue)
		return nil
	}

	// Try if the target field's type is convertible from the value's type
	if valueType.ConvertibleTo(field.Type()) {
		field.Set(valueValue.Convert(field.Type()))
		return nil
	}

	// If we reach here, the value couldn't be set
	return errors.Wrapf(rerr.NotSupported,
		"cannot set value of type %T to field of type %s", value, field.Type())
}

// handleInterfaceSlice handles conversion from []any to a target slice type
func handleInterfaceSlice(field reflect.Value, values []any) error {
	if field.Kind() != reflect.Slice {
		return errors.Wrapf(rerr.NotSupported,
			"cannot set []any to non-slice field of type %s", field.Type())
	}

	// Create a new slice of the appropriate type
	elemType := field.Type().Elem()
	slice := reflect.MakeSlice(field.Type(), 0, len(values))

	// Convert each element and append to the slice
	for _, val := range values {
		elem := reflect.New(elemType).Elem()
		if err := Set(elem, val); err != nil {
			return errors.Wrapf(err,
				"failed to convert element of []any to %s", elemType)
		}
		slice = reflect.Append(slice, elem)
	}

	field.Set(slice)
	return nil
}
