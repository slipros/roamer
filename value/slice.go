// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// SliceOption is a functional option for configuring SetSliceString behavior
type SliceOption func(*sliceOptions)

// sliceOptions contains configuration for SetSliceString
type sliceOptions struct {
	separator string
}

// defaultSliceOptions returns the default slice options
func defaultSliceOptions() sliceOptions {
	return sliceOptions{
		separator: ",",
	}
}

// WithSeparator sets a custom separator for joining string slices
func WithSeparator(sep string) SliceOption {
	return func(o *sliceOptions) {
		o.separator = sep
	}
}

// SetSliceString converts a string slice to the appropriate type for the target field
// and sets the field's value. This function handles conversion to various types,
// including joining strings into a single string, converting to []interface{},
// handling custom string slice types, numeric slices, and pointers to slices.
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - arr: The string slice to convert and set.
//   - options: Optional settings like separator for joining strings (default: ",").
//
// Returns:
//   - error: An error if the conversion or setting fails, or nil if successful.
//
// Example usage (internal to the package):
//
//	// Convert and set a string slice to a string field (joins with commas)
//	stringField := reflect.ValueOf(&myStruct).Elem().FieldByName("Tags")
//	if err := SetSliceString(stringField, []string{"tag1", "tag2", "tag3"}); err != nil {
//	    return err
//	}
//
//	// Convert and set a string slice to an []int field
//	intSliceField := reflect.ValueOf(&myStruct).Elem().FieldByName("IDs")
//	if err := SetSliceString(intSliceField, []string{"1", "2", "3"}); err != nil {
//	    return err // Will set [1, 2, 3]
//	}
//
//	// Convert with custom separator
//	stringField := reflect.ValueOf(&myStruct).Elem().FieldByName("Tags")
//	if err := SetSliceString(stringField, []string{"tag1", "tag2"}, WithSeparator("|")); err != nil {
//	    return err // Will set "tag1|tag2"
//	}
func SetSliceString(field reflect.Value, arr []string, options ...SliceOption) error {
	// Check if the field is settable
	if !field.CanSet() {
		return errors.Errorf("field of type %s is not settable", field.Type())
	}

	// Default options
	opts := defaultSliceOptions()
	for _, opt := range options {
		opt(&opts)
	}

	// Handle nil pointer initialization
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return SetSliceString(field.Elem(), arr, options...)
	}

	fieldType := field.Type()
	switch field.Kind() {
	case reflect.String:
		// Join the string slice into a single string with specified separator
		field.SetString(strings.Join(arr, opts.separator))
		return nil

	case reflect.Slice:
		elemType := fieldType.Elem()

		// Handle string slices (including custom types)
		if elemType.Kind() == reflect.String {
			// Create a new slice of the correct type and convert each element
			slice := reflect.MakeSlice(fieldType, len(arr), len(arr))
			for i, v := range arr {
				slice.Index(i).Set(reflect.ValueOf(v).Convert(elemType))
			}
			field.Set(slice)
			return nil
		}

		// Handle numeric slice types
		if isNumericKind(elemType.Kind()) {
			slice := reflect.MakeSlice(fieldType, 0, len(arr))
			for _, strVal := range arr {
				elemValue := reflect.New(elemType).Elem()
				if err := SetString(elemValue, strVal); err != nil {
					return errors.Wrapf(err, "failed to convert string '%s' to %s", strVal, elemType.String())
				}
				slice = reflect.Append(slice, elemValue)
			}
			field.Set(slice)
			return nil
		}

		// Handle boolean slices
		if elemType.Kind() == reflect.Bool {
			slice := reflect.MakeSlice(fieldType, 0, len(arr))
			for _, strVal := range arr {
				elemValue := reflect.New(elemType).Elem()
				if err := SetString(elemValue, strVal); err != nil {
					return errors.Wrapf(err, "failed to convert string '%s' to bool", strVal)
				}
				slice = reflect.Append(slice, elemValue)
			}
			field.Set(slice)
			return nil
		}

		// Handle []interface{} or any slice that accepts strings
		if elemType.Kind() == reflect.Interface {
			slice := reflect.MakeSlice(fieldType, len(arr), len(arr))
			for i, v := range arr {
				slice.Index(i).Set(reflect.ValueOf(v))
			}
			field.Set(slice)
			return nil
		}

	case reflect.Interface:
		// For interface{} fields, prefer setting as []string directly
		field.Set(reflect.ValueOf(arr))
		return nil
	}

	// If the field doesn't match any of the above types, return an error
	return errors.Wrapf(rerr.NotSupported,
		"cannot convert []string to field of type %s", fieldType.String())
}

// isNumericKind returns true if the kind is a numeric type
func isNumericKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
