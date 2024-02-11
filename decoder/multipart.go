package decoder

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/value"
)

const (
	// ContentTypeMultipartFormData content-type header for multipart form-data decoder.
	ContentTypeMultipartFormData = "multipart/form-data"
	// multipartFormDataMaxMemory max memory used by multipart form-data decoder for body parsing.
	defaultMultipartFormDataMaxMemory int64 = 32 << 20 // 32 MB
	tagValueAllFiles                        = ",allfiles"
	tagValueMultipartFormData               = "multipart"
)

// MultipartFormDataOptionsFunc function for setting options.
type MultipartFormDataOptionsFunc = func(*MultipartFormData)

// WithMaxMemory sets max memory.
func WithMaxMemory(maxMemory int64) MultipartFormDataOptionsFunc {
	return func(m *MultipartFormData) {
		m.maxMemory = maxMemory
	}
}

// MultipartFormData multipart form-data decoder.
type MultipartFormData struct {
	contentType string
	maxMemory   int64
}

// NewMultipartFormData returns new multipart form-data decoder.
func NewMultipartFormData(opts ...MultipartFormDataOptionsFunc) *MultipartFormData {
	m := MultipartFormData{
		contentType: ContentTypeMultipartFormData,
		maxMemory:   defaultMultipartFormDataMaxMemory,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

// Decode decodes url form value from http request into ptr.
//
// ptr must be pointer to a struct.
func (m *MultipartFormData) Decode(r *http.Request, ptr any) error {
	if err := r.ParseMultipartForm(m.maxMemory); err != nil {
		return errors.WithMessage(err, "parse multipart form")
	}

	v := reflect.Indirect(reflect.ValueOf(ptr))

	switch v.Kind() {
	case reflect.Struct:
		return m.parseStruct(r, v)
	default:
		return rerr.NotSupported
	}
}

// ContentType returns content type of url form decoder.
func (m *MultipartFormData) ContentType() string {
	return m.contentType
}

// setContentType set content-type value.
func (m *MultipartFormData) setContentType(contentType string) {
	m.contentType = contentType
}

func (m *MultipartFormData) parseStruct(r *http.Request, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		if !fieldType.IsExported() || len(fieldType.Tag) == 0 {
			continue
		}

		tagValue, ok := fieldType.Tag.Lookup(tagValueMultipartFormData)
		if !ok {
			continue
		}

		if len(r.Form) > 0 {
			if formValue, ok := m.parseFormValue(r.Form, tagValue); ok {
				if err := value.Set(v.Field(i), formValue); err != nil {
					return errors.WithMessagef(err, "set `%s` value to field `%s`", formValue, fieldType.Name)
				}

				continue
			}
		}

		switch tagValue {
		case tagValueAllFiles:
			files := make(MultipartFiles, 0, len(r.MultipartForm.File))
			for k := range r.MultipartForm.File {
				file, header, err := r.FormFile(k)
				if err != nil {
					return errors.WithMessagef(err, "parse form file for key %q", k)
				}

				files = append(files, MultipartFile{
					Key:    k,
					File:   file,
					Header: header,
				})
			}

			if err := m.setFileValue(v.Field(i), files); err != nil {
				return errors.WithMessagef(err, "set `%s` multipart value to field `%s`",
					tagValue, fieldType.Name)
			}
		default:
			files := r.MultipartForm.File[tagValue]
			if len(files) == 0 {
				continue
			}

			file, header, err := r.FormFile(tagValue)
			if err != nil {
				return errors.WithMessagef(err, "parse form file for key %q", tagValue)
			}

			multipartFile := MultipartFile{
				Key:    tagValue,
				File:   file,
				Header: header,
			}

			if err := m.setFileValue(v.Field(i), &multipartFile); err != nil {
				return errors.WithMessagef(err, "set `%s` multipart value to field `%s`",
					tagValue, fieldType.Name)
			}
		}
	}

	return nil
}

func (m *MultipartFormData) parseFormValue(form url.Values, tagValue string) (any, bool) {
	values, ok := form[tagValue]
	if !ok {
		return nil, false
	}

	if len(values) == 1 {
		return values[0], true
	}

	return values, true
}

func (m *MultipartFormData) setFileValue(field reflect.Value, value any) error {
	if field.Kind() == reflect.Pointer && field.IsNil() {
		// init ptr
		field.Set(reflect.New(field.Type().Elem()))
		field = reflect.Indirect(field)
	}

	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Pointer {
		// deref ptr
		valueType = valueType.Elem()
	}

	if field.Type().AssignableTo(valueType) {
		field.Set(reflect.Indirect(reflect.ValueOf(value)))
		return nil
	}

	return rerr.NotSupported
}
