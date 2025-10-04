// Package roamer provides a flexible HTTP request parser for Go applications.
// It extracts data from various parts of an HTTP request (headers, query parameters,
// cookies, body) into Go structures using struct tags, simplifying API development.
package roamer

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/slipros/assign"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/internal/cache"
	"github.com/slipros/roamer/parser"
	"golang.org/x/exp/maps"
)

const (
	parserCacheCapacity = 5
	maxParserCacheLen   = parserCacheCapacity * 6
)

// AfterParser is an interface that can be implemented by the target struct
// to execute custom logic after the HTTP request has been parsed.
//
//go:generate mockery --name=AfterParser --outpkg=mockroamer --output=./mockroamer
type AfterParser interface {
	// AfterParse is called after the HTTP request has been successfully parsed.
	// This method can be used to perform additional validation, data transformation,
	// or business logic based on the parsed data.
	AfterParse(r *http.Request) error
}

// RequireStructureCache is an interface for components that require
// a structure cache for efficient field analysis and caching.
//
// Components implementing this interface will receive a structure cache
// instance during Roamer initialization, allowing them to optimize
// reflection operations by caching struct field metadata.
//
// This interface is typically implemented by decoders that need to
// perform repetitive struct field analysis for the same types.
type RequireStructureCache interface {
	// SetStructureCache provides the component with a structure cache instance.
	// This method is called once during Roamer initialization to pass
	// the cache to components that need it for performance optimization.
	//
	// Parameters:
	//   - cache: The structure cache instance for storing field metadata.
	SetStructureCache(cache *cache.Structure)
}

// AssignExtensions is an optional interface that can be implemented by parsers and decoders
// to provide custom value assignment extensions. These extensions allow for specialized
// handling of complex types during the assignment process.
//
// When a parser or decoder implements this interface, its extension functions are
// automatically registered with the assignment system during Roamer initialization.
// This enables custom type conversions and value transformations beyond the standard
// assignment capabilities.
//
// Example implementation:
//
//	func (p *CustomParser) AssignExtensions() []assign.ExtensionFunc {
//	    return []assign.ExtensionFunc{
//	        func(value any) (func(to reflect.Value) error, bool) {
//	            // Custom assignment logic for specific types
//	            if customType, ok := value.(MyCustomType); ok {
//	                return func(to reflect.Value) error {
//	                    // Convert and assign the custom type
//	                    return assign.String(to, customType.String())
//	                }, true
//	            }
//	            return nil, false
//	        },
//	    }
//	}
type AssignExtensions interface {
	// AssignExtensions returns a slice of extension functions that provide
	// custom value assignment capabilities for specific types.
	//
	// Each ExtensionFunc takes a value and returns:
	//   - An assignment function that handles the conversion, if the extension
	//     can process the given value type
	//   - A boolean indicating whether this extension can handle the value
	//
	// Returns:
	//   A slice of extension functions for custom type assignments.
	AssignExtensions() []assign.ExtensionFunc
}

// Parse is a generic function that extracts data from an HTTP request into a value of type T.
// This is a convenience wrapper around the Roamer.Parse method that returns the parsed value
// directly instead of requiring a pointer parameter.
//
// The function creates a zero value of type T, parses the request data into it,
// and returns both the result and any error that occurred during parsing.
//
// Example:
//
//	type UserData struct {
//	    ID        int    `query:"id"`
//	    Name      string `json:"name"`
//	    UserAgent string `header:"User-Agent"`
//	}
//
//	// Parse request data directly into a value
//	userData, err := roamer.Parse[UserData](roamer, request)
//	if err != nil {
//	    return err
//	}
//	// Use userData...
//
// Parameters:
//   - r: The configured Roamer instance to use for parsing.
//   - req: The HTTP request to parse data from.
//
// Returns:
//   - T: The parsed data structure of the specified type.
//   - error: An error if parsing fails, or nil if successful.
func Parse[T any](r *Roamer, req *http.Request) (T, error) {
	if r == nil {
		var result T
		return result, errors.Wrap(rerr.NilValue, "roamer")
	}

	var result T
	err := r.Parse(req, &result)
	return result, err
}

// Roamer is a flexible HTTP request parser that extracts data from various parts
// of an HTTP request into Go structures using struct tags.
type Roamer struct {
	parsers                Parsers                // Collection of registered parsers
	decoders               Decoders               // Collection of registered decoders
	formatters             Formatters             // Collection of registered formatters
	reflectValueFormatters ReflectValueFormatters // Collection of registered reflectValueFormatters

	skipFilled    bool // Whether to skip fields that are already filled
	hasParsers    bool // Whether any parsers are registered
	hasDecoders   bool // Whether any decoders are registered
	hasFormatters bool // Whether any formatters are registered
	preserveBody  bool // Whether to preserve the request body after decoding

	parserCachePool sync.Pool
	structureCache  *cache.Structure

	assignExtensions []assign.ExtensionFunc
}

// NewRoamer creates a configured Roamer instance with optional configuration.
//
// Example:
//
//	// Basic Roamer with JSON decoder and query parser
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(decoder.NewJSON()),
//	    roamer.WithParsers(parser.NewQuery()),
//	)
//
//	// Roamer with multiple components
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(decoder.NewJSON(), decoder.NewFormURL()),
//	    roamer.WithParsers(parser.NewQuery(), parser.NewHeader()),
//	    roamer.WithFormatters(formatter.NewString()),
//	    roamer.WithSkipFilled(false), // Parse all fields, even if not zero
//	)
func NewRoamer(opts ...OptionsFunc) *Roamer {
	r := Roamer{
		parsers:                make(Parsers),
		decoders:               make(Decoders),
		formatters:             make(Formatters),
		reflectValueFormatters: make(ReflectValueFormatters),
		skipFilled:             true,
		parserCachePool: sync.Pool{
			New: func() any {
				return make(map[string]any, parserCacheCapacity)
			},
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	r.hasParsers = len(r.parsers) > 0
	r.hasDecoders = len(r.decoders) > 0
	r.hasFormatters = len(r.formatters) > 0

	r.structureCache = cache.NewStructure(
		cache.WithDecoders(r.decoders.Tags()),
		cache.WithParsers(maps.Keys(r.parsers)),
		cache.WithFormatters(maps.Keys(r.formatters)),
		cache.WithReflectValueFormatters(maps.Keys(r.reflectValueFormatters)),
	)

	for _, d := range r.decoders {
		if i, ok := d.(RequireStructureCache); ok {
			i.SetStructureCache(r.structureCache)
		}
	}

	return &r
}

// Parse extracts data from an HTTP request into the provided pointer (struct, slice, array, or map).
// For structs, it processes both the request body and other parts (headers, query parameters, cookies)
// according to struct tags. For slices, arrays, and maps, only the request body is processed.
//
// The target can implement AfterParser to execute custom logic after parsing is complete.
//
// Example:
//
//	type UserData struct {
//	    ID        int       `query:"id"`
//	    Name      string    `json:"name"`
//	    UserAgent string    `header:"User-Agent"`
//	}
//
//	var userData UserData
//	err := roamer.Parse(request, &userData)
func (r *Roamer) Parse(req *http.Request, ptr any) error {
	if req == nil {
		return errors.Wrapf(rerr.NilValue, "request")
	}

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

	if !r.hasParsers && !r.hasFormatters {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(ptr))
	t := v.Type()

	parserCache := r.parserCachePool.Get().(parser.Cache)
	defer func() {
		if len(parserCache) >= maxParserCacheLen {
			parserCache = make(parser.Cache, parserCacheCapacity)
		} else {
			clear(parserCache)
		}

		r.parserCachePool.Put(parserCache)
	}()

	fields := r.structureCache.Fields(t)

	for i := range fields {
		f := &fields[i]

		if !f.HasFormatters && !f.HasParsers && !f.HasDefault {
			continue
		}

		fieldValue := v.Field(f.Index)
		isZero := fieldValue.IsZero()

		if r.skipFilled && !isZero {
			if f.HasFormatters {
				if err := r.applyFormatters(f, fieldValue); err != nil {
					return errors.WithMessagef(err, "format field `%s` in struct `%T`", f.Name, ptr)
				}
			}

			continue
		}

		var parsedSuccessfully bool
		if f.HasParsers {
			for _, parserName := range f.Parsers {
				p, ok := r.parsers[parserName]
				if !ok {
					continue
				}

				parsedValue, ok := p.Parse(req, f.StructField.Tag, parserCache)
				if !ok {
					continue
				}

				if err := assign.Value(fieldValue, parsedValue, r.assignExtensions...); err != nil {
					return errors.Wrapf(err, "set `%s` value to field `%s` from tag `%s` for struct `%T`",
						parsedValue, f.Name, parserName, ptr)
				}

				parsedSuccessfully = true

				break
			}
		}

		if !parsedSuccessfully && f.HasDefault && isZero {
			if err := assign.Value(fieldValue, f.DefaultValue, r.assignExtensions...); err != nil {
				return errors.Wrapf(err, "set default value for field `%s`", f.Name)
			}
		}

		if f.HasFormatters {
			if err := r.applyFormatters(f, fieldValue); err != nil {
				return errors.WithMessagef(err, "format field `%s` in struct `%T`", f.Name, ptr)
			}
		}
	}

	return nil
}

func (r *Roamer) applyFormatters(field *cache.Field, fieldValue reflect.Value) error {
	for _, name := range field.ReflectValueFormatters {
		f, ok := r.reflectValueFormatters[name]
		if !ok {
			continue
		}

		if err := f.FormatReflectValue(field.StructField.Tag, fieldValue); err != nil {
			return err
		}
	}

	ptr, ok := assign.Pointer(fieldValue)
	if !ok {
		return nil
	}

	for _, name := range field.Formatters {
		f, ok := r.formatters[name]
		if !ok {
			continue
		}

		if err := f.Format(field.StructField.Tag, ptr); err != nil {
			return err
		}
	}

	return nil
}

// parseBody extracts data from the HTTP request body into the provided pointer
// using the appropriate decoder based on the request's Content-Type header.
// If preserveBody is enabled, it reads the entire body into memory and restores
// it after decoding so downstream handlers can read it again.
func (r *Roamer) parseBody(req *http.Request, ptr any) error {
	if !r.hasDecoders || req.Body == nil || req.ContentLength == 0 || req.Method == http.MethodGet {
		return nil
	}

	contentType := req.Header.Get("Content-Type")
	if idx := strings.IndexByte(contentType, ';'); idx != -1 {
		contentType = contentType[:idx]
	}

	d, ok := r.decoders[contentType]
	if !ok {
		return nil
	}

	// If body preservation is enabled, read the entire body into memory
	// and restore it after decoding
	if r.preserveBody {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return errors.Wrap(err, "read request body for preservation")
		}

		// Close the original body
		_ = req.Body.Close()

		// Replace body with a reader for decoding
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Decode directly from the request
		if err := d.Decode(req, ptr); err != nil {
			// Restore the body even on error so it can be inspected
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			return errors.WithStack(rerr.DecodeError{
				Err: errors.WithMessagef(err, "decode `%s` request body for `%T`", contentType, ptr),
			})
		}

		// Restore the body for downstream handlers
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		return nil
	}

	// Standard decoding without preservation
	if err := d.Decode(req, ptr); err != nil {
		return errors.WithStack(rerr.DecodeError{
			Err: errors.WithMessagef(err, "decode `%s` request body for `%T`", contentType, ptr),
		})
	}

	return nil
}
