package value

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// typeString is a reflect.Type for the string type.
// It's used for type comparison and conversion.
var typeString = reflect.TypeOf("")

// SetString converts a string value to the appropriate type for a target field.
// Handles conversion to numeric types, booleans, slices, time.Time and other types.
// Supports types implementing TextUnmarshaler or BinaryUnmarshaler interfaces.
//
// Parameters:
//   - field: Target field to set (reflect.Value).
//   - str: String value to convert and set.
//
// Returns:
//   - error: If conversion or assignment fails.
func SetString(field reflect.Value, str string) error {
	// Check if field can be set
	if !field.CanSet() {
		return errors.Wrapf(rerr.NotSupported, "field cannot be set")
	}

	// Special handling for empty strings - use zero values
	if str == "" {
		return handleEmptyString(field)
	}

	switch field.Kind() {
	case reflect.String:
		// Direct string assignment
		field.SetString(str)
		return nil

	case reflect.Bool:
		return setBoolFromString(field, str)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setIntFromString(field, str)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUintFromString(field, str)

	case reflect.Float32, reflect.Float64:
		return setFloatFromString(field, str)

	case reflect.Complex64, reflect.Complex128:
		return setComplexFromString(field, str)

	case reflect.Slice:
		return setSliceFromString(field, str)

	case reflect.Interface:
		// For interface{} fields, just set the string value
		field.Set(reflect.ValueOf(str))
		return nil

	case reflect.Ptr:
		// For pointer fields, initialize if nil and call recursively
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return SetString(field.Elem(), str)

	case reflect.Struct:
		return setStructFromString(field, str)

	case reflect.Map:
		return setMapFromString(field, str)
	}

	// Try to use custom unmarshalers for other types
	return tryUnmarshalers(field, str)
}

// handleEmptyString processes an empty string value based on the field type.
// For most types, it sets the zero value. For pointers to primitive types,
// it sets nil.
func handleEmptyString(field reflect.Value) error {
	switch field.Kind() {
	case reflect.String:
		// Empty string is a valid string value
		field.SetString("")
		return nil

	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		// For numeric and bool types, empty string means zero value
		field.Set(reflect.Zero(field.Type()))
		return nil

	case reflect.Slice:
		// For slices, don't append anything, keep it as is
		return nil

	case reflect.Ptr:
		// For pointers, set to nil (zero value)
		field.Set(reflect.Zero(field.Type()))
		return nil

	case reflect.Interface:
		// For interface, set to empty string
		field.Set(reflect.ValueOf(""))
		return nil
	}

	// Try unmarshalers with empty string
	return tryUnmarshalers(field, "")
}

// setBoolFromString converts a string to a boolean and sets the field value.
func setBoolFromString(field reflect.Value, str string) error {
	// Handle common string boolean representations
	str = strings.ToLower(str)
	switch str {
	case "1", "t", "true", "yes", "y", "on":
		field.SetBool(true)
		return nil
	case "0", "f", "false", "no", "n", "off":
		field.SetBool(false)
		return nil
	}

	// Use standard parsing as fallback
	parsed, err := strconv.ParseBool(str)
	if err != nil {
		return errors.Wrapf(err, "cannot convert string '%s' to bool", str)
	}
	field.SetBool(parsed)
	return nil
}

// setIntFromString converts a string to an integer and sets the field value.
// It handles decimal, hexadecimal (0x prefix), and octal (0 prefix) formats.
func setIntFromString(field reflect.Value, str string) error {
	// Determine base automatically (0 means auto-detect: 0x for hex, 0 for octal)
	base := 10
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		base = 0 // Auto-detect will use base 16 for 0x prefix
	} else if len(str) > 1 && strings.HasPrefix(str, "0") {
		base = 0 // Auto-detect will use base 8 for 0 prefix
	}

	// Get bit size based on field type
	bitSize := 0 // Default for int
	switch field.Kind() {
	case reflect.Int8:
		bitSize = 8
	case reflect.Int16:
		bitSize = 16
	case reflect.Int32:
		bitSize = 32
	case reflect.Int64:
		bitSize = 64
	}

	// Parse the string
	parsed, err := strconv.ParseInt(str, base, bitSize)
	if err != nil {
		return errors.Wrapf(err, "cannot convert string '%s' to %s", str, field.Kind())
	}

	// Set the value
	field.SetInt(parsed)
	return nil
}

// setUintFromString converts a string to an unsigned integer and sets the field value.
// It handles decimal, hexadecimal (0x prefix), and octal (0 prefix) formats.
func setUintFromString(field reflect.Value, str string) error {
	// Determine base automatically
	base := 10
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		base = 0 // Auto-detect will use base 16 for 0x prefix
	} else if len(str) > 1 && strings.HasPrefix(str, "0") {
		base = 0 // Auto-detect will use base 8 for 0 prefix
	}

	// Get bit size based on field type
	bitSize := 0 // Default for uint
	switch field.Kind() {
	case reflect.Uint8:
		bitSize = 8
	case reflect.Uint16:
		bitSize = 16
	case reflect.Uint32:
		bitSize = 32
	case reflect.Uint64:
		bitSize = 64
	}

	// Parse the string
	parsed, err := strconv.ParseUint(str, base, bitSize)
	if err != nil {
		return errors.Wrapf(err, "cannot convert string '%s' to %s", str, field.Kind())
	}

	// Set the value
	field.SetUint(parsed)
	return nil
}

// setFloatFromString converts a string to a float and sets the field value.
func setFloatFromString(field reflect.Value, str string) error {
	// Get bit size based on field type
	bitSize := 64 // Default for float64
	if field.Kind() == reflect.Float32 {
		bitSize = 32
	}

	// Parse the string
	parsed, err := strconv.ParseFloat(str, bitSize)
	if err != nil {
		return errors.Wrapf(err, "cannot convert string '%s' to %s", str, field.Kind())
	}

	// Set the value
	field.SetFloat(parsed)
	return nil
}

// setComplexFromString converts a string to a complex number and sets the field value.
func setComplexFromString(field reflect.Value, str string) error {
	// Get bit size based on field type
	bitSize := 128 // Default for complex128
	if field.Kind() == reflect.Complex64 {
		bitSize = 64
	}

	// Parse the string
	parsed, err := strconv.ParseComplex(str, bitSize)
	if err != nil {
		return errors.Wrapf(err, "cannot convert string '%s' to %s", str, field.Kind())
	}

	// Set the value
	field.SetComplex(parsed)
	return nil
}

// setSliceFromString handles conversion from string to various slice types.
func setSliceFromString(field reflect.Value, str string) error {
	elemType := field.Type().Elem()

	switch elemType.Kind() {
	case reflect.Uint8:
		// []byte/[]uint8 - direct conversion
		field.SetBytes([]byte(str))
		return nil

	case reflect.String:
		if !strings.Contains(str, ",") {
			// Single value case - avoid slice creation
			strValue := reflect.ValueOf(str)
			if elemType != typeString && strValue.Type().ConvertibleTo(elemType) {
				strValue = strValue.Convert(elemType)
			}
			field.Set(reflect.Append(field, strValue))
			return nil
		}

		// Pre-calculate capacity to reduce allocations
		commaCount := strings.Count(str, ",")
		parts := make([]string, 0, commaCount+1)

		// Use strings.FieldsFunc for better performance
		parts = strings.FieldsFunc(str, func(r rune) bool {
			return r == ','
		})

		// Pre-allocate slice with known capacity
		newSlice := reflect.MakeSlice(field.Type(), 0, len(parts))

		for _, part := range parts {
			if part = strings.TrimSpace(part); part != "" {
				strValue := reflect.ValueOf(part)
				if elemType != typeString && strValue.Type().ConvertibleTo(elemType) {
					strValue = strValue.Convert(elemType)
				}
				newSlice = reflect.Append(newSlice, strValue)
			}
		}
		field.Set(newSlice)

		return nil

	default:
		// For other slice types, try to parse the string as a comma-separated list
		if strings.Contains(str, ",") {
			parts := strings.Split(str, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					// Create a new element for the slice
					newElem := reflect.New(elemType).Elem()

					// Try to set the new element with the string value
					if err := SetString(newElem, trimmed); err == nil {
						field.Set(reflect.Append(field, newElem))
					} else {
						return errors.Wrapf(err, "cannot convert '%s' to element of %s", trimmed, field.Type())
					}
				}
			}
			return nil
		}

		// Try to set a single element
		newElem := reflect.New(elemType).Elem()
		if err := SetString(newElem, str); err == nil {
			field.Set(reflect.Append(field, newElem))
			return nil
		}

		return errors.Wrapf(rerr.NotSupported, "cannot convert string to %s", field.Type())
	}
}

// setStructFromString handles conversion from string to struct types.
// Currently supports time.Time directly.
func setStructFromString(field reflect.Value, str string) error {
	// Special handling for time.Time
	if field.Type() == typeTime {
		return setTimeFromString(field, str)
	}

	// For other structs, try using unmarshalers
	return tryUnmarshalers(field, str)
}

// setTimeFromString parses a string into a time.Time value using various layouts.
func setTimeFromString(field reflect.Value, str string) error {
	t, err := parseTime(str)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(t))

	return nil
}

// setMapFromString attempts to set a map value from a string.
// Currently, this only supports maps with string keys and a simple format like "key1:value1,key2:value2".
func setMapFromString(field reflect.Value, str string) error {
	// Only support maps with string keys for now
	if field.Type().Key().Kind() != reflect.String {
		return errors.Wrapf(rerr.NotSupported, "only maps with string keys are supported for string conversion")
	}

	// Check if the map needs to be initialized
	if field.IsNil() {
		field.Set(reflect.MakeMap(field.Type()))
	}

	// Parse key-value pairs (format: "key1:value1,key2:value2")
	if !strings.Contains(str, ":") {
		return errors.Wrapf(rerr.NotSupported, "invalid map format, expected 'key:value' pairs separated by commas")
	}

	pairs := strings.Split(str, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			return errors.Errorf("invalid map key-value pair: %s", pair)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Create a new value for the map
		elemType := field.Type().Elem()
		elem := reflect.New(elemType).Elem()

		// Set the value
		if err := SetString(elem, value); err != nil {
			return errors.Wrapf(err, "cannot convert '%s' to map value type %s", value, elemType)
		}

		// Set the key-value pair in the map
		field.SetMapIndex(reflect.ValueOf(key), elem)
	}

	return nil
}

// tryUnmarshalers attempts to use TextUnmarshaler or BinaryUnmarshaler interfaces
// to unmarshal the string into the field.
func tryUnmarshalers(field reflect.Value, str string) error {
	// For types that implement TextUnmarshaler or BinaryUnmarshaler,
	// we need to get a pointer to the field
	if !field.CanAddr() {
		return errors.WithStack(rerr.NotSupported)
	}

	ptr := field.Addr()
	if !ptr.CanInterface() {
		return errors.WithStack(rerr.NotSupported)
	}

	return implementsBytesUnmarshaler(ptr.Interface(), str)
}

// implementsBytesUnmarshaler uses TextUnmarshaler or BinaryUnmarshaler interfaces
// to unmarshal a string into a value, if supported.
func implementsBytesUnmarshaler(ptr any, str string) error {
	switch i := ptr.(type) {
	case encoding.TextUnmarshaler:
		// For types like time.Time that implement TextUnmarshaler
		if err := i.UnmarshalText([]byte(str)); err != nil {
			return errors.Wrapf(err, "TextUnmarshaler failed for '%s'", str)
		}
		return nil

	case encoding.BinaryUnmarshaler:
		// For types that implement BinaryUnmarshaler
		if err := i.UnmarshalBinary([]byte(str)); err != nil {
			return errors.Wrapf(err, "BinaryUnmarshaler failed for '%s'", str)
		}
		return nil
	}

	return errors.Wrapf(rerr.NotSupported, "type %T does not implement UnmarshalText or UnmarshalBinary", ptr)
}
