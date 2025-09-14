// Package value provides utilities for type conversion and setting values in Go structs.
// It is used by the roamer package to convert parsed values from HTTP requests
// into appropriate types for struct fields.
package value

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// Set assigns a value to a reflect.Value field with automatic type conversion.
// Handles initialization of nil pointers, various primitive types, and interfaces.
//
// Key features:
// - Converts between compatible types (string to int, float to string, etc.)
// - Initializes nil pointers when needed
// - Supports fmt.Stringer interface
// - Handles slices and common primitive types
//
// Parameters:
//   - field: Target field to set (reflect.Value).
//   - value: Value to assign (any type).
//
// Returns:
//   - error: If value could not be set due to type incompatibility.
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

	kind := field.Kind()

	// Initialize and dereference nil pointers recursively
	if kind == reflect.Pointer {
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
		if kind == reflect.Bool {
			field.SetBool(t)
			return nil
		}
		return SetString(field, strconv.FormatBool(t))
	case *bool:
		if t == nil {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		if kind == reflect.Bool {
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
		if kind == reflect.Slice {
			return handleInterfaceSlice(field, t)
		}
	}

	// Handle http.Cookie objects specially to extract just the value
	// This allows cookie values to be converted to appropriate types (int, string, etc.)
	// while maintaining backward compatibility for code that expects *http.Cookie
	if cookie, ok := value.(*http.Cookie); ok {
		return SetString(field, cookie.Value)
	}

	// Handle types that implement fmt.Stringer
	if stringer, ok := value.(fmt.Stringer); ok {
		return SetString(field, stringer.String())
	}

	// Handle general assignable types
	valueValue := reflect.ValueOf(value)
	valueType := valueValue.Type()

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

	fieldType := field.Type()

	// Check if the value's type can be assigned to the target field's type
	if valueType.AssignableTo(fieldType) {
		field.Set(valueValue)

		return nil
	}

	// Try if the value's type is convertible to the target field's type
	if valueType.ConvertibleTo(fieldType) {
		// Perform safe conversion with panic recovery
		var convertErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					convertErr = errors.Errorf("conversion panic: %v", r)
				}
			}()

			convertedValue := valueValue.Convert(fieldType)
			field.Set(convertedValue)
		}()

		if convertErr != nil {
			return convertErr
		}

		return nil
	}

	// If we reach here, the value couldn't be set
	return errors.Wrapf(rerr.NotSupported, "cannot set value of type %T to field of type %s", value, fieldType)
}

// handleInterfaceSlice handles conversion from []any to a target slice type.
// This function creates a new slice of the appropriate target type and converts
// each element from the interface{} slice to the target element type.
//
// Parameters:
//   - field: The target slice field to populate (must be of slice kind).
//   - values: Slice of interface{} values to convert and assign.
//
// Returns:
//   - error: An error if conversion fails for any element, or nil if successful.
func handleInterfaceSlice(field reflect.Value, values []any) error {
	if field.Kind() != reflect.Slice {
		return errors.Wrapf(rerr.NotSupported, "cannot set []any to non-slice field of type %s", field.Type())
	}

	// Create a new slice of the appropriate type
	elemType := field.Type().Elem()
	slice := reflect.MakeSlice(field.Type(), len(values), len(values))

	// Convert each element and set directly by index
	for i, val := range values {
		elem := reflect.New(elemType).Elem()
		if err := Set(elem, val); err != nil {
			return errors.Wrapf(err, "failed to convert element of []any to %s", elemType)
		}

		slice.Index(i).Set(elem)
	}

	field.Set(slice)

	return nil
}
