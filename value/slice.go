// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

var (
	// typeAnySlice is a reflect.Type for []any.
	// It's used for type comparison and conversion.
	typeAnySlice = reflect.TypeOf([]any{})

	// typeStringSlice is a reflect.Type for []string.
	// It's used for type comparison and conversion.
	typeStringSlice = reflect.TypeOf([]string{})
)

// SetSliceString converts a string slice to the appropriate type for the target field
// and sets the field's value. This function handles conversion to various types,
// including joining strings into a single string, converting to []interface{},
// and handling custom string slice types.
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - arr: The string slice to convert and set.
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
//	// Convert and set a string slice to another string slice field
//	sliceField := reflect.ValueOf(&myStruct).Elem().FieldByName("Categories")
//	if err := SetSliceString(sliceField, []string{"cat1", "cat2", "cat3"}); err != nil {
//	    return err
//	}
func SetSliceString(field reflect.Value, arr []string) error {
	fieldType := field.Type()
	switch field.Kind() {
	case reflect.String:
		// Join the string slice into a single string with comma separator
		field.SetString(strings.Join(arr, ","))
		return nil

	case reflect.Slice:
		elemType := fieldType.Elem()
		switch elemType.Kind() {
		case reflect.String:
			// Handle custom string slice types (e.g., type Tags []string)
			if fieldType != typeStringSlice && typeString.ConvertibleTo(elemType) {
				// Create a new slice of the correct type and convert each element
				slice := reflect.MakeSlice(fieldType, 0, len(arr))
				for _, v := range arr {
					casted := reflect.ValueOf(v).Convert(elemType)
					slice = reflect.Append(slice, casted)
				}
				field.Set(slice)
				return nil
			}

			// Direct assignment for []string
			field.Set(reflect.ValueOf(arr))
			return nil

		case reflect.Interface:
			// Handle []interface{} fields
			if field.Type().AssignableTo(typeAnySlice) {
				// Convert string slice to []interface{}
				s := make([]any, 0, len(arr))
				for _, v := range arr {
					s = append(s, v)
				}
				field.Set(reflect.ValueOf(s))
				return nil
			}
		}

	case reflect.Interface:
		// Handle interface{} fields that might expect a string slice
		// Note: This is a special case and might need reconsideration
		if v, ok := field.Interface().([]string); ok {
			field.Set(reflect.ValueOf(v))

			return nil
		}
	}

	// If the field doesn't match any of the above types, return an error
	return errors.WithStack(rerr.NotSupported)
}
