package decoder

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	rerr "github.com/SLIpros/roamer/err"
	"github.com/SLIpros/roamer/value"
)

const (
	// ContentTypeFormURL content-type header for url form decoder.
	ContentTypeFormURL = "application/x-www-form-urlencoded"
	// SplitSymbol array split symbol.
	SplitSymbol     = ","
	tagValueFormURL = "form"
)

// FormURLOptionsFunc function for setting options.
type FormURLOptionsFunc func(*FormURL)

// WithDisabledSplit disables array splitting.
func WithDisabledSplit() FormURLOptionsFunc {
	return func(f *FormURL) {
		f.split = false
	}
}

// WithSplitSymbol sets array split symbol.
func WithSplitSymbol(splitSymbol string) FormURLOptionsFunc {
	return func(f *FormURL) {
		f.splitSymbol = splitSymbol
	}
}

// FormURL url form decoder.
type FormURL struct {
	split       bool
	splitSymbol string
}

// NewFormURL returns new url form decoder.
func NewFormURL(opts ...FormURLOptionsFunc) *FormURL {
	f := FormURL{split: true, splitSymbol: SplitSymbol}

	for _, opt := range opts {
		opt(&f)
	}

	return &f
}

// Decode decodes url form value from http request into ptr.
//
// ptr must have a type of either struct or map.
func (f *FormURL) Decode(r *http.Request, ptr any) error {
	if err := r.ParseForm(); err != nil {
		return errors.WithMessage(err, "parse http form")
	}

	v := reflect.Indirect(reflect.ValueOf(ptr))
	t := v.Type()

	switch v.Kind() {
	case reflect.Struct:
		return f.parseStruct(&v, t, r.PostForm)
	case reflect.Map:
		return f.parseMap(&v, t, r.PostForm)
	default:
		return rerr.NotSupported
	}
}

// ContentType returns content type of url form decoder.
func (f *FormURL) ContentType() string {
	return ContentTypeFormURL
}

func (f *FormURL) parseFormValue(form url.Values, tag reflect.StructTag) (any, bool) {
	tagValue, ok := tag.Lookup(tagValueFormURL)
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

func (f *FormURL) parseStruct(v *reflect.Value, t reflect.Type, form url.Values) error {
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
		return rerr.NotSupported
	}

	mValue := t.Elem()
	switch mValue.Kind() {
	case reflect.String:
		m := make(map[string]string, len(form))
		for k, v := range form {
			if len(v) == 1 || !f.split {
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

	return rerr.NotSupported
}
