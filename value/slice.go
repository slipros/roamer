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

// WithSeparator sets a custom separator for joining string slices.
// Used as an option in SetSliceString.
//
// Example: WithSeparator("|") // Join with pipe character
func WithSeparator(sep string) SliceOption {
	return func(o *sliceOptions) {
		o.separator = sep
	}
}

// SetSliceString converts a string slice to appropriate types for target fields.
// Handles conversions to strings (joins elements), numeric slices, boolean slices,
// and other compatible types.
//
// Parameters:
//   - field: Target field to set (reflect.Value).
//   - arr: String slice to convert and set.
//   - options: Optional settings like custom separator (default: ",").
//
// Returns:
//   - error: If conversion or assignment fails.
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
