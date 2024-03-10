package value

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

var typeSliceOfAny = reflect.TypeOf([]any{})

// SetSliceString sets slice of strings into a field.
func SetSliceString(field reflect.Value, arr []string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(strings.Join(arr, ","))
		return nil
	case reflect.Slice:
		elemKind := field.Type().Elem().Kind()
		switch elemKind {
		case reflect.String:
			field.Set(reflect.ValueOf(arr))
			return nil
		case reflect.Interface:
			if field.Type().AssignableTo(typeSliceOfAny) {
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
