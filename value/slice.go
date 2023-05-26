package value

import (
	"reflect"
	"strings"

	roamerError "github.com/SLIpros/roamer/error"
)

// SetSliceString set slice of strings to field.
func SetSliceString(field *reflect.Value, arr []string) error {
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
			s := make([]any, 0, len(arr))
			for _, v := range arr {
				s = append(s, v)
			}

			field.Set(reflect.ValueOf(s))
			return nil
		}
	case reflect.Interface:
		field.Set(reflect.ValueOf(arr))
		return nil
	}

	return roamerError.ErrNotSupported
}
