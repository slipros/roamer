package decoder

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/slipros/exp"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/value"
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
	contentType                 string
	split                       bool
	splitSymbol                 string
	experimentalFastStructField bool
}

// NewFormURL returns new url form decoder.
func NewFormURL(opts ...FormURLOptionsFunc) *FormURL {
	f := FormURL{contentType: ContentTypeFormURL, split: true, splitSymbol: SplitSymbol}

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

// EnableExperimentalFastStructFieldParser enables the use of experimental fast struct field parser.
func (f *FormURL) EnableExperimentalFastStructFieldParser() {
	f.experimentalFastStructField = true
}

// ContentType returns content-type header value.
func (f *FormURL) ContentType() string {
	return f.contentType
}

// setContentType set content-type value.
func (f *FormURL) setContentType(contentType string) {
	f.contentType = contentType
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

func (f *FormURL) parseStruct(v *reflect.Value, t reflect.Type, form url.Values) (err error) {
	var fieldType reflect.StructField
	for i := 0; i < v.NumField(); i++ {
		if f.experimentalFastStructField {
			fieldType, err = exp.FastStructField(v, i)
			if err != nil {
				return errors.WithStack(err)
			}
		} else {
			fieldType = t.Field(i)
		}

		if !fieldType.IsExported() || len(fieldType.Tag) == 0 {
			continue
		}

		formValue, ok := f.parseFormValue(form, fieldType.Tag)
		if !ok {
			continue
		}

		if err := value.Set(v.Field(i), formValue); err != nil {
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
