// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"encoding"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// typeString is a reflect.Type for the string type.
// It's used for type comparison and conversion.
var typeString = reflect.TypeOf("")

// SetString converts a string value to the appropriate type for the target field
// and sets the field's value. This function handles conversion to various types,
// including all numeric types, booleans, slices, and types that implement the
// TextUnmarshaler or BinaryUnmarshaler interfaces.
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - str: The string value to convert and set.
//
// Returns:
//   - error: An error if the conversion or setting fails, or nil if successful.
//
// Example usage (internal to the package):
//
//	// Convert and set a string value to an int field
//	intField := reflect.ValueOf(&myStruct).Elem().FieldByName("Count")
//	if err := SetString(intField, "42"); err != nil {
//	    return err
//	}
//
//	// Convert and set a string value to a time.Time field (via TextUnmarshaler)
//	timeField := reflect.ValueOf(&myStruct).Elem().FieldByName("CreatedAt")
//	if err := SetString(timeField, "2023-01-15T12:00:00Z"); err != nil {
//	    return err
//	}
func SetString(field reflect.Value, str string) error {
	switch field.Kind() {
	case reflect.String:
		// Direct string assignment
		field.SetString(str)
		return nil

	case reflect.Bool:
		// Convert string to bool
		parsed, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		field.SetBool(parsed)
		return nil

	case reflect.Int8:
		// Convert string to int8
		parsed, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
		return nil

	case reflect.Int16:
		// Convert string to int16
		parsed, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
		return nil

	case reflect.Int32:
		// Convert string to int32
		parsed, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
		return nil

	case reflect.Int64:
		// Convert string to int64
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
		return nil

	case reflect.Int:
		// Convert string to int
		parsed, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
		return nil

	case reflect.Uint8:
		// Convert string to uint8
		parsed, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return err
		}
		field.SetUint(parsed)
		return nil

	case reflect.Uint16:
		// Convert string to uint16
		parsed, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return err
		}
		field.SetUint(parsed)
		return nil

	case reflect.Uint32:
		// Convert string to uint32
		parsed, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return err
		}
		field.SetUint(parsed)
		return nil

	case reflect.Uint64:
		// Convert string to uint64
		parsed, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(parsed)
		return nil

	case reflect.Uint:
		// Convert string to uint
		parsed, err := strconv.ParseUint(str, 10, 0)
		if err != nil {
			return err
		}
		field.SetUint(parsed)
		return nil

	case reflect.Float32:
		// Convert string to float32
		parsed, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return err
		}
		field.SetFloat(parsed)
		return nil

	case reflect.Float64:
		// Convert string to float64
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		field.SetFloat(parsed)
		return nil

	case reflect.Complex64:
		// Convert string to complex64
		parsed, err := strconv.ParseComplex(str, 64)
		if err != nil {
			return err
		}
		field.SetComplex(parsed)
		return nil

	case reflect.Complex128:
		// Convert string to complex128
		parsed, err := strconv.ParseComplex(str, 128)
		if err != nil {
			return err
		}
		field.SetComplex(parsed)
		return nil

	case reflect.Slice:
		// Handle slices specially
		elemType := field.Type().Elem()
		switch elemType.Kind() {
		case reflect.Uint8:
			// []byte/[]uint8 - convert string to byte slice
			field.SetBytes([]byte(str))
			return nil
		case reflect.String:
			// []string - append string to slice
			strValue := reflect.ValueOf(str)
			if elemType != typeString && strValue.Type().ConvertibleTo(elemType) {
				strValue = strValue.Convert(elemType)
			}
			field.Set(reflect.Append(field, strValue))
			return nil
		}

	case reflect.Interface:
		// For interface{} fields, just set the string value
		field.Set(reflect.ValueOf(str))
		return nil

	case reflect.Ptr:
		// For pointer fields, dereference and call recursively
		return SetString(field.Elem(), str)
	}

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

// implementsBytesUnmarshaler checks if the provided value implements
// TextUnmarshaler or BinaryUnmarshaler interfaces, and if so, calls
// the appropriate method to unmarshal the string.
//
// Parameters:
//   - ptr: A pointer to the value that might implement the unmarshaler interfaces.
//   - str: The string value to unmarshal.
//
// Returns:
//   - error: An error if unmarshaling fails or if the value doesn't implement
//     either interface, or nil if successful.
func implementsBytesUnmarshaler(ptr any, str string) error {
	switch i := ptr.(type) {
	case encoding.TextUnmarshaler:
		// For types like time.Time that implement TextUnmarshaler
		return i.UnmarshalText([]byte(str))
	case encoding.BinaryUnmarshaler:
		// For types that implement BinaryUnmarshaler
		return i.UnmarshalBinary([]byte(str))
	}

	return errors.WithStack(rerr.NotSupported)
}
