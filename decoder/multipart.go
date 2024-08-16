package decoder

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/pkg/errors"
	"github.com/slipros/exp"
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

// MultipartFormDataOptionsFunc function for setting multipart options.
type MultipartFormDataOptionsFunc = func(*MultipartFormData)

// WithMaxMemory sets max memory.
func WithMaxMemory(maxMemory int64) MultipartFormDataOptionsFunc {
	return func(m *MultipartFormData) {
		m.maxMemory = maxMemory
	}
}

// MultipartFormData multipart form-data decoder.
type MultipartFormData struct {
	contentType                 string
	skipFilled                  bool
	maxMemory                   int64
	experimentalFastStructField bool
}

// NewMultipartFormData returns new multipart form-data decoder.
func NewMultipartFormData(opts ...MultipartFormDataOptionsFunc) *MultipartFormData {
	m := MultipartFormData{
		contentType: ContentTypeMultipartFormData,
		skipFilled:  true,
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
		return m.parseStruct(r, &v)
	default:
		return errors.WithStack(rerr.NotSupported)
	}
}

// EnableExperimentalFastStructFieldParser enables the use of experimental fast struct field parser.
func (m *MultipartFormData) EnableExperimentalFastStructFieldParser() {
	m.experimentalFastStructField = true
}

// ContentType returns content type of url form decoder.
func (m *MultipartFormData) ContentType() string {
	return m.contentType
}

// setContentType set content-type value.
func (m *MultipartFormData) setContentType(contentType string) {
	m.contentType = contentType
}

// setSkipFilled sets skip filled value.
func (m *MultipartFormData) setSkipFilled(skip bool) {
	m.skipFilled = skip
}

// parseStruct parses structure from http request into a ptr.
func (m *MultipartFormData) parseStruct(r *http.Request, v *reflect.Value) (err error) {
	t := v.Type()
	var fieldType reflect.StructField

	for i := range v.NumField() {
		if m.experimentalFastStructField {
			ft, exists := exp.FastStructField(v, i)
			if !exists {
				// should never happen - anomaly.
				return errors.WithStack(rerr.FieldIndexOutOfBounds)
			}

			fieldType = ft
		} else {
			fieldType = t.Field(i)
		}

		if !fieldType.IsExported() || len(fieldType.Tag) == 0 {
			continue
		}

		tagValue, ok := fieldType.Tag.Lookup(tagValueMultipartFormData)
		if !ok {
			continue
		}

		if len(r.Form) > 0 {
			if formValue, ok := m.parseFormValue(r.Form, tagValue); ok {
				fieldValue := v.Field(i)
				if m.skipFilled && !fieldValue.IsZero() {
					continue
				}

				if err := value.Set(fieldValue, formValue); err != nil {
					return errors.WithMessagef(err, "set `%s` value to field `%s`", formValue, fieldType.Name)
				}

				continue
			}
		}

		if r.MultipartForm == nil {
			continue
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
			if files := r.MultipartForm.File[tagValue]; len(files) == 0 {
				continue
			}

			file, header, err := r.FormFile(tagValue)
			if err != nil {
				return errors.WithMessagef(err, "parse form file for key %q", tagValue)
			}

			fieldValue := v.Field(i)
			if m.skipFilled && !fieldValue.IsZero() {
				continue
			}

			multipartFile := MultipartFile{
				Key:    tagValue,
				File:   file,
				Header: header,
			}

			if err := m.setFileValue(fieldValue, &multipartFile); err != nil {
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
	fieldType := field.Type()

	if fieldType.AssignableTo(valueType) {
		field.Set(reflect.ValueOf(value))

		return nil
	}

	if valueType.Kind() == reflect.Pointer && fieldType.AssignableTo(valueType.Elem()) {
		// deref ptr
		field.Set(reflect.Indirect(reflect.ValueOf(value)))

		return nil
	}

	return errors.WithStack(rerr.NotSupported)
}
