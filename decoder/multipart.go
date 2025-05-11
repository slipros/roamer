// Package decoder provides decoders for extracting data from HTTP request bodies.
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
	// ContentTypeMultipartFormData is the Content-Type header value for multipart form data.
	// This is used to match requests with the appropriate decoder.
	ContentTypeMultipartFormData = "multipart/form-data"

	// defaultMultipartFormDataMaxMemory is the default maximum memory in bytes
	// that will be used to parse multipart form data. Content beyond this limit
	// will be stored in temporary files.
	defaultMultipartFormDataMaxMemory int64 = 32 << 20 // 32 MB

	// tagValueAllFiles is a special tag value that instructs the parser
	// to collect all files in the request.
	tagValueAllFiles = ",allfiles"

	// tagValueMultipartFormData is the struct tag name used for multipart form values.
	tagValueMultipartFormData = "multipart"
)

// MultipartFormDataOptionsFunc is a function type for configuring a MultipartFormData decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type MultipartFormDataOptionsFunc = func(*MultipartFormData)

// WithMaxMemory sets the maximum memory in bytes that will be used to parse multipart form data.
// Content beyond this limit will be stored in temporary files.
// The default value is 32 MB.
//
// Example:
//
//	// Create a multipart form decoder with a 10 MB memory limit
//	multipartDecoder := decoder.NewMultipartFormData(
//	    decoder.WithMaxMemory(10 << 20), // 10 MB
//	)
func WithMaxMemory(maxMemory int64) MultipartFormDataOptionsFunc {
	return func(m *MultipartFormData) {
		m.maxMemory = maxMemory
	}
}

// MultipartFormData is a decoder for handling multipart form data,
// including file uploads. It can parse both form fields and uploaded files
// into struct fields based on struct tags.
type MultipartFormData struct {
	contentType                 string // The Content-Type header value that this decoder handles
	skipFilled                  bool   // Whether to skip fields that are already filled
	maxMemory                   int64  // The maximum memory in bytes to use for parsing
	experimentalFastStructField bool   // Whether to use experimental fast struct field access
}

// NewMultipartFormData creates a new MultipartFormData decoder with the specified options.
// By default, it handles requests with Content-Type "multipart/form-data",
// skips fields that are already filled, and uses a 32 MB memory limit.
//
// Example:
//
//	// Create a multipart form decoder with default settings
//	multipartDecoder := decoder.NewMultipartFormData()
//
//	// Create a multipart form decoder with custom options
//	multipartDecoder := decoder.NewMultipartFormData(
//	    decoder.WithMaxMemory(10 << 20), // 10 MB memory limit
//	    decoder.WithSkipFilled(false),   // Don't skip fields that are already filled
//	)
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(multipartDecoder),
//	)
//
//	// Example struct using multipart tags
//	type UploadRequest struct {
//	    Title       string              `multipart:"title"`
//	    Description string              `multipart:"description"`
//	    Avatar      *decoder.MultipartFile `multipart:"avatar"` // Single file
//	    Gallery     decoder.MultipartFiles `multipart:",allfiles"` // All files
//	}
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

// Decode parses multipart form data from an HTTP request into the provided pointer.
// The pointer must be to a struct.
//
// The decoder handles both form fields and file uploads. Form fields are parsed
// into regular struct fields, while file uploads are parsed into MultipartFile
// or MultipartFiles fields.
//
// Parameters:
//   - r: The HTTP request containing the multipart form data to decode.
//   - ptr: A pointer to the target struct where the decoded data will be stored.
//
// Returns:
//   - error: An error if decoding fails, or nil if successful.
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

// EnableExperimentalFastStructFieldParser enables the use of an experimental
// fast struct field parser. This can improve performance but may not be
// as stable as the standard parser.
//
// This method is part of the internal Experiment interface and is primarily
// used by the roamer package.
func (m *MultipartFormData) EnableExperimentalFastStructFieldParser() {
	m.experimentalFastStructField = true
}

// ContentType returns the Content-Type header value that this decoder handles.
// For the MultipartFormData decoder, this is "multipart/form-data" by default.
// This method is used by the roamer package to match requests with the appropriate decoder.
func (m *MultipartFormData) ContentType() string {
	return m.contentType
}

// setContentType sets the Content-Type header value that this decoder handles.
// This is primarily used internally by option functions.
func (m *MultipartFormData) setContentType(contentType string) {
	m.contentType = contentType
}

// setSkipFilled sets whether the decoder should skip fields that are already filled.
// This is primarily used internally by option functions.
func (m *MultipartFormData) setSkipFilled(skip bool) {
	m.skipFilled = skip
}

// parseStruct parses multipart form data into a struct.
// It handles both form fields and file uploads.
//
// Parameters:
//   - r: The HTTP request containing the multipart form data.
//   - v: A pointer to the reflect.Value of the struct.
//
// Returns:
//   - error: An error if parsing fails, or nil if successful.
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

// parseFormValue extracts a form value from the provided form data.
// It handles both single values and multiple values.
//
// Parameters:
//   - form: The form data to extract values from.
//   - tagValue: The form field name.
//
// Returns:
//   - any: The extracted form value (string or []string).
//   - bool: Whether a value was found.
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

// setFileValue sets a file value to a field.
// It handles both single files and collections of files.
//
// Parameters:
//   - field: The target field to set (as a reflect.Value).
//   - value: The file value to set (MultipartFile, *MultipartFile, or MultipartFiles).
//
// Returns:
//   - error: An error if the value could not be set, or nil if successful.
func (m *MultipartFormData) setFileValue(field reflect.Value, value any) error {
	if field.Kind() == reflect.Pointer && field.IsNil() {
		// Initialize nil pointers
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
		// Dereference pointers
		field.Set(reflect.Indirect(reflect.ValueOf(value)))

		return nil
	}

	return errors.WithStack(rerr.NotSupported)
}
