package decoder

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/internal/cache"
	"github.com/slipros/roamer/value"
	"golang.org/x/exp/slices"
)

const (
	// ContentTypeMultipartFormData is the Content-Type header value for multipart form data.
	// This is used to match requests with the appropriate decoder.
	ContentTypeMultipartFormData = "multipart/form-data"

	// TagMultipart is the struct tag name used for multipart form values.
	TagMultipart = "multipart"

	// defaultMultipartFormDataMaxMemory is the default maximum memory in bytes
	// that will be used to parse multipart form data. Content beyond this limit
	// will be stored in temporary files.
	defaultMultipartFormDataMaxMemory int64 = 32 << 20 // 32 MB

	// tagValueAllFiles is a special tag value that instructs the parser
	// to collect all files in the request.
	tagValueAllFiles = ",allfiles"
)

// MultipartFormDataOptionsFunc is a function type for configuring a MultipartFormData decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type MultipartFormDataOptionsFunc = func(*MultipartFormData)

// WithMaxMemory sets the maximum memory limit for parsing multipart data.
// Content beyond this limit will be stored in temporary files.
// Default is 32 MB.
//
// Example: decoder.WithMaxMemory(10 << 20) // 10 MB
func WithMaxMemory(maxMemory int64) MultipartFormDataOptionsFunc {
	return func(m *MultipartFormData) {
		m.maxMemory = maxMemory
	}
}

// MultipartFormData is a decoder for handling multipart form data,
// including file uploads. It can parse both form fields and uploaded files
// into struct fields based on struct tags.
type MultipartFormData struct {
	contentType string // The Content-Type header value that this decoder handles
	skipFilled  bool   // Whether to skip fields that are already filled
	maxMemory   int64  // The maximum memory in bytes to use for parsing

	structureCache *cache.StructureCache
}

// NewMultipartFormData creates a decoder for multipart/form-data content,
// including file uploads. Handles both form fields and uploaded files.
//
// Example:
//
//	// Default settings (32MB memory limit)
//	multipartDecoder := decoder.NewMultipartFormData()
//
//	// Custom memory limit
//	multipartDecoder := decoder.NewMultipartFormData(
//	    decoder.WithMaxMemory(10 << 20), // 10 MB
//	)
//
//	// Example struct
//	type UploadRequest struct {
//	    Title   string              `multipart:"title"`
//	    Avatar  *decoder.MultipartFile `multipart:"avatar"` // Single file
//	    Gallery decoder.MultipartFiles `multipart:",allfiles"` // All files
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

// Decode parses multipart form data into a struct pointer.
// Handles both form fields and file uploads.
//
// Parameters:
//   - r: The HTTP request with multipart form data.
//   - ptr: Target struct pointer.
//
// Returns:
//   - error: Error if decoding fails, nil if successful.
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

// ContentType returns the Content-Type header value that this decoder handles.
// For the MultipartFormData decoder, this is "multipart/form-data" by default.
// This method is used by the roamer package to match requests with the appropriate decoder.
func (m *MultipartFormData) ContentType() string {
	return m.contentType
}

// Tag returns the struct tag name used for multipart field mapping.
// For the MultipartFormData decoder, this is "multipart" by default.
func (m *MultipartFormData) Tag() string {
	return TagMultipart
}

// SetStructureCache assigns a structure cache to the decoder for improved performance.
// The cache stores precomputed field information to avoid reflection overhead on each request.
func (m *MultipartFormData) SetStructureCache(cache *cache.StructureCache) {
	m.structureCache = cache
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

	if m.structureCache != nil {
		fields := m.structureCache.Fields(t)
		for i := range fields {
			f := &fields[i]

			if len(f.Decoders) == 0 || !slices.Contains(f.Decoders, TagMultipart) {
				continue
			}

			tagValue, ok := f.StructField.Tag.Lookup(TagMultipart)
			if !ok {
				continue
			}

			if err := m.parseField(r, v, f.Index, f.Name, tagValue); err != nil {
				return err
			}
		}

		return nil
	}

	for i := range v.NumField() {
		f := t.Field(i)

		if !f.IsExported() || len(f.Tag) == 0 {
			continue
		}

		tagValue, ok := f.Tag.Lookup(TagMultipart)
		if !ok {
			continue
		}

		if err := m.parseField(r, v, i, f.Name, tagValue); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultipartFormData) parseField(
	r *http.Request,
	v *reflect.Value,
	fieldIndex int,
	fieldName string,
	tagValue string,
) error {
	if len(r.Form) > 0 {
		if formValue, ok := m.parseFormValue(r.Form, tagValue); ok {
			fieldValue := v.Field(fieldIndex)
			if m.skipFilled && !fieldValue.IsZero() {
				return nil
			}

			if err := value.Set(fieldValue, formValue); err != nil {
				return errors.WithMessagef(err, "set `%s` value to field `%s`", formValue, fieldName)
			}

			return nil
		}
	}

	if r.MultipartForm == nil {
		return nil
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

		if err := m.setFileValue(v.Field(fieldIndex), files); err != nil {
			return errors.WithMessagef(err, "set `%s` multipart value to field `%s`",
				tagValue, fieldName)
		}
	default:
		if files := r.MultipartForm.File[tagValue]; len(files) == 0 {
			return nil
		}

		file, header, err := r.FormFile(tagValue)
		if err != nil {
			return errors.WithMessagef(err, "parse form file for key %q", tagValue)
		}

		fieldValue := v.Field(fieldIndex)
		if m.skipFilled && !fieldValue.IsZero() {
			return nil
		}

		multipartFile := MultipartFile{
			Key:    tagValue,
			File:   file,
			Header: header,
		}

		if err := m.setFileValue(fieldValue, &multipartFile); err != nil {
			return errors.WithMessagef(err, "set `%s` multipart value to field `%s`",
				tagValue, fieldName)
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
