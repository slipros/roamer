package value

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

var (
	typeAnySlice    = reflect.TypeOf([]any{})
	typeStringSlice = reflect.TypeOf([]string{})
)

// SetSliceString sets slice of strings into a field.
func SetSliceString(field reflect.Value, arr []string) error {
	fieldType := field.Type()
	switch field.Kind() {
	case reflect.String:
		field.SetString(strings.Join(arr, ","))

		return nil
	case reflect.Slice:
		elemType := fieldType.Elem()
		switch elemType.Kind() {
		case reflect.String:
			if fieldType != typeStringSlice && typeString.ConvertibleTo(elemType) {
				slice := reflect.MakeSlice(fieldType, 0, len(arr))
				for _, v := range arr {
					casted := reflect.ValueOf(v).Convert(elemType)
					slice = reflect.Append(slice, casted)
				}

				field.Set(slice)

				return nil
			}

			field.Set(reflect.ValueOf(arr))

			return nil
		case reflect.Interface:
			if field.Type().AssignableTo(typeAnySlice) {
				s := make([]any, 0, len(arr))
				for _, v := range arr {
					s = append(s, v)
				}

				field.Set(reflect.ValueOf(s))

				return nil
			}
		}
	case reflect.Interface:
		// FIXME: make any assignable
		//nolint:gocritic // no other way
		switch field.Interface().(type) {
		case []string:
			field.Set(reflect.ValueOf(arr))

			return nil
		}
	}

	return errors.WithStack(rerr.NotSupported)
}
