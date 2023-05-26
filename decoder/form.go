package decoder

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	roamerError "github.com/SLIpros/roamer/error"
	"github.com/SLIpros/roamer/value"
)

const (
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

type FormURLEncoded struct {
	splitSymbol string
}

func NewFormURLEncoded(splitSymbol string) *FormURLEncoded {
	return &FormURLEncoded{splitSymbol: splitSymbol}
}

func (f *FormURLEncoded) ContentType() string {
	return ContentTypeFormURLEncoded
}

func (f *FormURLEncoded) Decode(r *http.Request, ptr any) error {
	if err := r.ParseForm(); err != nil {
		return errors.WithMessage(err, "parse http form")
	}

	v := reflect.Indirect(reflect.ValueOf(ptr))
	t := v.Type()

	switch v.Kind() {
	case reflect.Struct:
		return f.parseStructure(&v, t, r.PostForm)
	case reflect.Map:
		return f.parseMap(&v, t, r.PostForm)
	default:
		return roamerError.ErrNotSupported
	}
}

func (f *FormURLEncoded) parseFormValue(form url.Values, tag reflect.StructTag) (any, bool) {
	tagValue, ok := tag.Lookup("form")
	if !ok {
		return nil, false
	}

	values, ok := form[tagValue]
	if !ok {
		return nil, false
	}

	if len(values) == 1 {
		return values[0], true
	}

	return values, true
}

func (f *FormURLEncoded) parseStructure(v *reflect.Value, t reflect.Type, form url.Values) error {
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		if !fieldType.IsExported() || len(fieldType.Tag) == 0 {
			continue
		}

		formValue, ok := f.parseFormValue(form, fieldType.Tag)
		if !ok {
			continue
		}

		fieldValue := v.Field(i)
		if err := value.Set(&fieldValue, formValue); err != nil {
			return errors.WithMessagef(err, "set `%s` value to field `%s`", formValue, fieldType.Name)
		}
	}

	return nil
}

func (f *FormURLEncoded) parseMap(v *reflect.Value, t reflect.Type, form url.Values) error {
	switch t.Key().Kind() {
	case reflect.String:
		{
			mValue := t.Elem()
			switch mValue.Kind() {
			case reflect.String:
				m := make(map[string]string, len(form))
				for k, v := range form {
					if len(v) == 1 {
						m[k] = v[0]
						continue
					}

					m[k] = strings.Join(v, f.splitSymbol)
				}

				v.Set(reflect.ValueOf(m))
				return nil
			case reflect.Interface:
				m := make(map[string]any, len(form))
				for k, v := range form {
					if len(v) == 1 {
						m[k] = v[0]
						continue
					}

					m[k] = v
				}

				v.Set(reflect.ValueOf(m))
				return nil
			case reflect.Slice:
				sType := mValue.Elem()
				switch sType.Kind() {
				case reflect.String:
					v.Set(reflect.ValueOf(form))
					return nil
				}
			}
		}
	}

	return roamerError.ErrNotSupported
}
