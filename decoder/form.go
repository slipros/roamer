package decoder

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/value"
)

const (
	// ContentTypeFormURL is the Content-Type header value for URL-encoded form data.
	// This is used to match requests with the appropriate decoder.
	ContentTypeFormURL = "application/x-www-form-urlencoded"

	// TagForm is the struct tag name used for URL-encoded form values.
	TagForm = "form"

	// SplitSymbol is the default character used to split form values when
	// multiple values are provided for the same field.
	SplitSymbol = ","
)

// FormURLOptionsFunc is a function type for configuring a FormURL decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type FormURLOptionsFunc func(*FormURL)

// WithDisabledSplit disables automatic splitting of form values into slices.
//
// Example: With splitting disabled, "tags=foo,bar,baz" will parse as a single
// string "foo,bar,baz" rather than a slice ["foo", "bar", "baz"]
func WithDisabledSplit() FormURLOptionsFunc {
	return func(f *FormURL) {
		f.split = false
	}
}

// WithSplitSymbol sets the character used for splitting form values (default: comma).
//
// Example: With splitSymbol set to ";", a form field "tags=foo;bar;baz"
// will parse as a slice ["foo", "bar", "baz"]
func WithSplitSymbol(splitSymbol string) FormURLOptionsFunc {
	return func(f *FormURL) {
		f.splitSymbol = splitSymbol
	}
}

// FormURL is a decoder for handling URL-encoded form data.
// It can parse form data into structs and maps, handling both
// single values and multiple values.
type FormURL struct {
	contentType string // The Content-Type header value that this decoder handles
	skipFilled  bool   // Whether to skip fields that are already filled
	split       bool   // Whether to split comma-separated values
	splitSymbol string // The character to use when splitting values
}

// NewFormURL creates a FormURL decoder that handles application/x-www-form-urlencoded content.
// By default, it skips already filled fields and splits comma-separated values.
//
// Example:
//
//	// Default form decoder
//	formDecoder := decoder.NewFormURL()
//
//	// Custom configuration
//	formDecoder := decoder.NewFormURL(
//	    decoder.WithDisabledSplit(),       // Don't split comma-separated values
//	    decoder.WithSplitSymbol(";"),      // Use semicolon as separator
//	)
//
//	// Example struct
//	type SearchRequest struct {
//	    Query string   `form:"q"`
//	    Tags  []string `form:"tags"` // Can handle "tags=foo,bar,baz"
//	}
func NewFormURL(opts ...FormURLOptionsFunc) *FormURL {
	f := FormURL{
		contentType: ContentTypeFormURL,
		skipFilled:  true,
		split:       true,
		splitSymbol: SplitSymbol,
	}

	for _, opt := range opts {
		opt(&f)
	}

	return &f
}

// Decode parses URL-encoded form data into a struct or map.
// For structs, uses the "form" tag to map fields.
// For maps, populates with form field names as keys.
//
// Parameters:
//   - r: The HTTP request with form data.
//   - ptr: Target pointer (struct or map).
//
// Returns:
//   - error: Error if decoding fails, nil if successful.
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
		return errors.WithStack(rerr.NotSupported)
	}
}

// ContentType returns the Content-Type header value that this decoder handles.
// For the FormURL decoder, this is "application/x-www-form-urlencoded" by default.
// This method is used by the roamer package to match requests with the appropriate decoder.
func (f *FormURL) ContentType() string {
	return f.contentType
}

// Tag returns the struct tag name used for form field mapping.
// For the FormURL decoder, this is "form" by default.
func (f *FormURL) Tag() string {
	return TagForm
}

// setContentType sets the Content-Type header value that this decoder handles.
// This is primarily used internally by option functions.
func (f *FormURL) setContentType(contentType string) {
	f.contentType = contentType
}

// setSkipFilled sets whether the decoder should skip fields that are already filled.
// This is primarily used internally by option functions.
func (f *FormURL) setSkipFilled(skip bool) {
	f.skipFilled = skip
}

// parseFormValue extracts a form value from the provided form data.
// It handles both single values and multiple values.
//
// Parameters:
//   - form: The form data to extract values from.
//   - tag: The struct tag containing the form field name.
//
// Returns:
//   - any: The extracted form value (string or []string).
//   - bool: Whether a value was found.
func (f *FormURL) parseFormValue(form url.Values, tag reflect.StructTag) (any, bool) {
	tagValue, ok := tag.Lookup(TagForm)
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

// parseStruct parses form data into a struct.
// It maps form field names to struct fields using the "form" tag.
//
// Parameters:
//   - v: A pointer to the reflect.Value of the struct.
//   - t: The reflect.Type of the struct.
//   - form: The form data to parse.
//
// Returns:
//   - error: An error if parsing fails, or nil if successful.
func (f *FormURL) parseStruct(v *reflect.Value, t reflect.Type, form url.Values) (err error) {
	var fieldType reflect.StructField
	for i := range v.NumField() {
		fieldType = t.Field(i)

		if !fieldType.IsExported() || len(fieldType.Tag) == 0 {
			continue
		}

		formValue, ok := f.parseFormValue(form, fieldType.Tag)
		if !ok {
			continue
		}

		fieldValue := v.Field(i)
		if f.skipFilled && !fieldValue.IsZero() {
			continue
		}

		if err := value.Set(fieldValue, formValue); err != nil {
			return errors.WithMessagef(err, "set `%s` value to field `%s`", formValue, fieldType.Name)
		}
	}

	return nil
}

// parseMap parses form data into a map.
// It populates the map with form field names as keys and field values as values.
//
// The function supports different map types:
//   - map[string]string: Single form values
//   - map[string]interface{}: Both single and multiple form values
//   - map[string][]string: Direct form data
//
// Parameters:
//   - v: A pointer to the reflect.Value of the map.
//   - t: The reflect.Type of the map.
//   - form: The form data to parse.
//
// Returns:
//   - error: An error if parsing fails, or nil if successful.
func (f *FormURL) parseMap(v *reflect.Value, t reflect.Type, form url.Values) error {
	if t.Key().Kind() != reflect.String {
		return errors.WithStack(rerr.NotSupported)
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

	return errors.WithStack(rerr.NotSupported)
}
