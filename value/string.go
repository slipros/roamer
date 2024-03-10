package value

import (
	"encoding"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// SetString sets string into a field.
func SetString(field reflect.Value, str string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(str)
		return nil
	case reflect.Bool:
		parsed, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}

		field.SetBool(parsed)
		return nil
	case reflect.Int8:
		parsed, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return err
		}

		field.SetInt(parsed)
		return nil
	case reflect.Int16:
		parsed, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return err
		}

		field.SetInt(parsed)
		return nil
	case reflect.Int32:
		parsed, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return err
		}

		field.SetInt(parsed)
		return nil
	case reflect.Int64:
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}

		field.SetInt(parsed)
		return nil
	case reflect.Int:
		parsed, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return err
		}

		field.SetInt(parsed)
		return nil
	case reflect.Uint8:
		parsed, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return err
		}

		field.SetUint(parsed)
		return nil
	case reflect.Uint16:
		parsed, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return err
		}

		field.SetUint(parsed)
		return nil
	case reflect.Uint32:
		parsed, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return err
		}

		field.SetUint(parsed)
		return nil
	case reflect.Uint64:
		parsed, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}

		field.SetUint(parsed)
		return nil
	case reflect.Uint:
		parsed, err := strconv.ParseUint(str, 10, 0)
		if err != nil {
			return err
		}

		field.SetUint(parsed)
		return nil
	case reflect.Float32:
		parsed, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return err
		}

		field.SetFloat(parsed)
		return nil
	case reflect.Float64:
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}

		field.SetFloat(parsed)
		return nil
	case reflect.Complex64:
		parsed, err := strconv.ParseComplex(str, 64)
		if err != nil {
			return err
		}

		field.SetComplex(parsed)
		return nil
	case reflect.Complex128:
		parsed, err := strconv.ParseComplex(str, 128)
		if err != nil {
			return err
		}

		field.SetComplex(parsed)
		return nil
	case reflect.Slice:
		elemKind := field.Type().Elem().Kind()
		switch elemKind {
		case reflect.Uint8:
			field.SetBytes([]byte(str))
			return nil
		case reflect.String:
			field.Set(reflect.Append(field, reflect.ValueOf(str)))
			return nil
		}
	case reflect.Interface:
		field.Set(reflect.ValueOf(str))
		return nil
	case reflect.Ptr:
		return SetString(field.Elem(), str)
	}

	if !field.CanAddr() {
		return errors.WithStack(rerr.NotSupported)
	}

	ptr := field.Addr()
	if !ptr.CanInterface() {
		return errors.WithStack(rerr.NotSupported)
	}

	return implementsBytesUnmarshaler(ptr.Interface(), str)
}

// implementsBytesUnmarshaler checks for interface implementation and calls it if there is a match.
func implementsBytesUnmarshaler(ptr any, str string) error {
	switch i := ptr.(type) {
	case encoding.TextUnmarshaler:
		return i.UnmarshalText([]byte(str))
	case encoding.BinaryUnmarshaler:
		return i.UnmarshalBinary([]byte(str))
	}

	return errors.WithStack(rerr.NotSupported)
}
