// Package decoder provides decoders for extracting data from HTTP request bodies.
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
	// ContentTypeFormURL is the Content-Type header value for URL-encoded form data.
	// This is used to match requests with the appropriate decoder.
	ContentTypeFormURL = "application/x-www-form-urlencoded"

	// SplitSymbol is the default character used to split form values when
	// multiple values are provided for the same field.
	SplitSymbol = ","

	// tagValueFormURL is the struct tag name used for URL-encoded form values.
	tagValueFormURL = "form"
)

// FormURLOptionsFunc is a function type for configuring a FormURL decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type FormURLOptionsFunc func(*FormURL)

// WithDisabledSplit disables the automatic splitting of form values.
// By default, if a field is set to handle multiple values (e.g., a slice),
// the decoder will split values using the split symbol (default: comma).
// This option disables that behavior.
//
// Example:
//
//	// Create a form decoder with splitting disabled
//	formDecoder := decoder.NewFormURL(decoder.WithDisabledSplit())
//
//	// With splitting disabled, a form field like "tags=foo,bar,baz" will be
//	// parsed as a single string "foo,bar,baz" rather than a slice ["foo", "bar", "baz"]
func WithDisabledSplit() FormURLOptionsFunc {
	return func(f *FormURL) {
		f.split = false
	}
}

// WithSplitSymbol sets the character used to split form values.
// By default, the decoder uses a comma (,) as the split symbol.
// This option allows using a different character instead.
//
// Example:
//
//	// Create a form decoder that splits on semicolons instead of commas
//	formDecoder := decoder.NewFormURL(decoder.WithSplitSymbol(";"))
//
//	// With this configuration, a form field like "tags=foo;bar;baz" will be
//	// parsed as a slice ["foo", "bar", "baz"]
func WithSplitSymbol(splitSymbol string) FormURLOptionsFunc {
	return func(f *FormURL) {
		f.splitSymbol = splitSymbol
	}
}

// FormURL is a decoder for handling URL-encoded form data.
// It can parse form data into structs and maps, handling both
// single values and multiple values.
type FormURL struct {
	contentType                 string // The Content-Type header value that this decoder handles
	skipFilled                  bool   // Whether to skip fields that are already filled
	split                       bool   // Whether to split comma-separated values
	splitSymbol                 string // The character to use when splitting values
	experimentalFastStructField bool   // Whether to use experimental fast struct field access
}

// NewFormURL creates a new FormURL decoder with the specified options.
// By default, it handles requests with Content-Type "application/x-www-form-urlencoded",
// skips fields that are already filled, and splits comma-separated values.
//
// Example:
//
//	// Create a form decoder with default settings
//	formDecoder := decoder.NewFormURL()
//
//	// Create a form decoder with custom options
//	formDecoder := decoder.NewFormURL(
//	    decoder.WithDisabledSplit(),         // Don't split comma-separated values
//	    decoder.WithSplitSymbol(";"),        // Use semicolon as separator (if splitting is enabled)
//	    decoder.WithSkipFilled(false),       // Don't skip fields that are already filled
//	)
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(formDecoder),
//	)
//
//	// Example struct using form tags
//	type SearchRequest struct {
//	    Query string   `form:"q"`
//	    Page  int      `form:"page"`
//	    Tags  []string `form:"tags"` // Can be provided as "tags=foo,bar,baz"
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

// Decode parses URL-encoded form data from an HTTP request into the provided pointer.
// The pointer must be to a struct or a map.
//
// For structs, the decoder uses the "form" tag to map form fields to struct fields.
// For maps, the decoder populates the map with form field names as keys.
//
// Parameters:
//   - r: The HTTP request containing the form data to decode.
//   - ptr: A pointer to the target value (struct or map) where the decoded data will be stored.
//
// Returns:
//   - error: An error if decoding fails, or nil if successful.
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

// EnableExperimentalFastStructFieldParser enables the use of an experimental
// fast struct field parser. This can improve performance but may not be
// as stable as the standard parser.
//
// This method is part of the internal Experiment interface and is primarily
// used by the roamer package.
func (f *FormURL) EnableExperimentalFastStructFieldParser() {
	f.experimentalFastStructField = true
}

// ContentType returns the Content-Type header value that this decoder handles.
// For the FormURL decoder, this is "application/x-www-form-urlencoded" by default.
// This method is used by the roamer package to match requests with the appropriate decoder.
func (f *FormURL) ContentType() string {
	return f.contentType
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
		if f.experimentalFastStructField {
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
