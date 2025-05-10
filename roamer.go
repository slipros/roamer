// Package roamer provides a flexible HTTP request parser for Go applications.
// It allows easy extraction of data from various parts of an HTTP request
// (headers, query parameters, cookies, body) into Go structures using struct tags.
package roamer

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/slipros/exp"
	rerr "github.com/slipros/roamer/err"
	rexp "github.com/slipros/roamer/internal/experiment"
	"github.com/slipros/roamer/parser"
	"github.com/slipros/roamer/value"
)

// AfterParser is an interface that can be implemented by the target struct
// to execute custom logic after the HTTP request has been parsed.
//
//go:generate mockery --name=AfterParser --outpkg=mock --output=./mock
type AfterParser interface {
	// AfterParse is called after the HTTP request has been successfully parsed.
	// This method can be used to perform additional validation, data transformation,
	// or business logic based on the parsed data.
	AfterParse(r *http.Request) error
}

// Roamer is a flexible HTTP request parser that extracts data from various parts
// of an HTTP request into Go structures using struct tags.
type Roamer struct {
	parsers                     Parsers    // Collection of registered parsers
	decoders                    Decoders   // Collection of registered decoders
	formatters                  Formatters // Collection of registered formatters
	skipFilled                  bool       // Whether to skip fields that are already filled
	hasParsers                  bool       // Whether any parsers are registered
	hasDecoders                 bool       // Whether any decoders are registered
	hasFormatters               bool       // Whether any formatters are registered
	experimentalFastStructField bool       // Whether to use experimental fast struct field access
}

// NewRoamer creates and returns a new configured Roamer instance.
// It accepts optional configuration functions to customize the behavior.
//
// Example:
//
//	// Create a basic Roamer with JSON decoder and query parser
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(decoder.NewJSON()),
//	    roamer.WithParsers(parser.NewQuery()),
//	)
//
//	// Create Roamer with multiple parsers and formatters
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(decoder.NewJSON(), decoder.NewFormURL()),
//	    roamer.WithParsers(
//	        parser.NewQuery(),
//	        parser.NewHeader(),
//	        parser.NewCookie(),
//	    ),
//	    roamer.WithFormatters(formatter.NewString()),
//	    roamer.WithSkipFilled(false), // Parse all fields, even if not zero
//	)
func NewRoamer(opts ...OptionsFunc) *Roamer {
	r := Roamer{
		parsers:    make(Parsers),
		decoders:   make(Decoders),
		formatters: make(Formatters),
		skipFilled: true,
	}

	for _, opt := range opts {
		opt(&r)
	}

	r.hasParsers = len(r.parsers) > 0
	r.hasDecoders = len(r.decoders) > 0
	r.hasFormatters = len(r.formatters) > 0

	if r.experimentalFastStructField {
		r.enableExperimentalFeatures()
	}

	return &r
}

// Parse extracts data from an HTTP request into the provided pointer.
// The pointer must be to a struct, slice, array, or map.
//
// If the pointer is to a struct, both the request body and other parts of the request
// (headers, query parameters, cookies) will be parsed according to the struct tags.
//
// If the pointer is to a slice, array, or map, only the request body will be parsed.
//
// The target type can implement the AfterParser interface to execute custom logic
// after parsing is complete.
//
// Example:
//
//	// Define a struct with tags for parsing
//	type UserData struct {
//	    ID        int       `query:"id"`
//	    Name      string    `json:"name"`
//	    UserAgent string    `header:"User-Agent"`
//	    SessionID string    `cookie:"session_id"`
//	}
//
//	// Parse request into the struct
//	var userData UserData
//	err := roamer.Parse(request, &userData)
func (r *Roamer) Parse(req *http.Request, ptr any) error {
	if ptr == nil {
		return errors.Wrapf(rerr.NilValue, "ptr")
	}

	t := reflect.TypeOf(ptr)
	if t.Kind() != reflect.Pointer {
		return errors.Wrapf(rerr.NotPtr, "`%T`", ptr)
	}

	switch t.Elem().Kind() {
	case reflect.Struct:
		if err := r.parseStruct(req, ptr); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if err := r.parseBody(req, ptr); err != nil {
			return err
		}
	default:
		return errors.Wrapf(rerr.NotSupported, "`%T`", ptr)
	}

	if p, ok := ptr.(AfterParser); ok {
		return p.AfterParse(req)
	}

	return nil
}

// parseStruct parses an HTTP request into a struct pointer by extracting data
// from the request body and other parts of the request (headers, query parameters, cookies)
// according to the struct tags.
func (r *Roamer) parseStruct(req *http.Request, ptr any) error {
	if err := r.parseBody(req, ptr); err != nil {
		return err
	}

	if !r.hasParsers {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(ptr))
	t := v.Type()

	var fieldType reflect.StructField

	fieldsAmount := v.NumField()
	cache := make(parser.Cache, fieldsAmount)

	for i := range fieldsAmount {
		if r.experimentalFastStructField {
			ft, exists := exp.FastStructField(&v, i)
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

		fieldValue := v.Field(i)
		if r.skipFilled && !fieldValue.IsZero() {
			if r.hasFormatters {
				if err := r.formatFieldValue(&fieldType, fieldValue); err != nil {
					return errors.WithMessagef(err, "format field `%s` in struct `%T`", fieldType.Name, ptr)
				}
			}

			continue
		}

		for tag, p := range r.parsers {
			parsedValue, ok := p.Parse(req, fieldType.Tag, cache)
			if !ok {
				continue
			}

			if err := value.Set(fieldValue, parsedValue); err != nil {
				return errors.Wrapf(err, "set `%s` value to field `%s` from tag `%s` for struct `%T`",
					parsedValue, fieldType.Name, tag, ptr)
			}

			break
		}

		if r.hasFormatters {
			if err := r.formatFieldValue(&fieldType, fieldValue); err != nil {
				return errors.WithMessagef(err, "format field `%s` in struct `%T`", fieldType.Name, ptr)
			}
		}
	}

	return nil
}

// formatFieldValue applies registered formatters to the field value if any formatter
// is applicable to the field's tags. This allows post-processing of parsed values
// (e.g., trimming strings, converting case, etc.).
func (r *Roamer) formatFieldValue(fieldType *reflect.StructField, fieldValue reflect.Value) error {
	if !r.formatters.has(fieldType.Tag) {
		return nil
	}

	fieldPtrValue, ok := value.Pointer(fieldValue)
	if !ok {
		return nil
	}

	for _, f := range r.formatters {
		if err := f.Format(fieldType.Tag, fieldPtrValue); err != nil {
			return err
		}
	}

	return nil
}

// parseBody extracts data from the HTTP request body into the provided pointer
// using the appropriate decoder based on the request's Content-Type header.
func (r *Roamer) parseBody(req *http.Request, ptr any) error {
	if !r.hasDecoders || req.ContentLength == 0 || req.Method == http.MethodGet {
		return nil
	}

	contentType := req.Header.Get("Content-Type")
	if base, _, found := strings.Cut(contentType, ";"); found {
		contentType = base
	}

	d, ok := r.decoders[contentType]
	if !ok {
		return nil
	}

	if err := d.Decode(req, ptr); err != nil {
		return errors.WithStack(rerr.DecodeError{
			Err: errors.WithMessagef(err, "decode `%s` request body for `%T`", contentType, ptr),
		})
	}

	return nil
}

// enableExperimentalFeatures configures experimental features in the registered decoders.
// Currently, this only enables the fast struct field parser if available.
func (r *Roamer) enableExperimentalFeatures() {
	for _, d := range r.decoders {
		e, ok := d.(rexp.Experiment)
		if !ok {
			continue
		}

		e.EnableExperimentalFastStructFieldParser()
	}
}
