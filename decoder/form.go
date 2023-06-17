package decoder

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	roamerError "github.com/SLIpros/roamer/err"
	"github.com/SLIpros/roamer/value"
)

const (
	// ContentTypeFormURL content-type header for url form decoder.
	ContentTypeFormURL = "application/x-www-form-urlencoded"
)

// FormURL url form decoder.
type FormURL struct {
	splitSymbol string
}

// NewFormURL returns new url form decoder.
func NewFormURL(splitSymbol string) *FormURL {
	return &FormURL{splitSymbol: splitSymbol}
}

// Decode decodes url form value from http request into ptr.
//
// Ptr must have a type of either struct or map.
func (f *FormURL) Decode(r *http.Request, ptr any) error {
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
		return roamerError.NotSupported
	}
}

// ContentType returns content type of url form decoder.
func (f *FormURL) ContentType() string {
	return ContentTypeFormURL
}

func (f *FormURL) parseFormValue(form url.Values, tag reflect.StructTag) (any, bool) {
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

func (f *FormURL) parseStructure(v *reflect.Value, t reflect.Type, form url.Values) error {
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

func (f *FormURL) parseMap(v *reflect.Value, t reflect.Type, form url.Values) error {
	if t.Key().Kind() != reflect.String {
		return roamerError.NotSupported
	}

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
		if sType.Kind() == reflect.String {
			v.Set(reflect.ValueOf(form))
			return nil
		}
	}

	return roamerError.NotSupported
}
