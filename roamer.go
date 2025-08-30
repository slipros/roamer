// Package roamer provides a flexible HTTP request parser for Go applications.
// It extracts data from various parts of an HTTP request (headers, query parameters,
// cookies, body) into Go structures using struct tags, simplifying API development.
package roamer

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/internal/cache"
	"github.com/slipros/roamer/parser"
	"github.com/slipros/roamer/value"
	"golang.org/x/exp/maps"
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
type RequireStructureCache interface {
	SetStructureCache(cache *cache.StructureCache)
}

// Roamer is a flexible HTTP request parser that extracts data from various parts
// of an HTTP request into Go structures using struct tags.
type Roamer struct {
	parsers       Parsers    // Collection of registered parsers
	decoders      Decoders   // Collection of registered decoders
	formatters    Formatters // Collection of registered formatters
	skipFilled    bool       // Whether to skip fields that are already filled
	hasParsers    bool       // Whether any parsers are registered
	hasDecoders   bool       // Whether any decoders are registered
	hasFormatters bool       // Whether any formatters are registered

	parserCachePool sync.Pool
	structureCache  *cache.StructureCache
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
		parsers:    make(Parsers),
		decoders:   make(Decoders),
		formatters: make(Formatters),
		skipFilled: true,
		parserCachePool: sync.Pool{
			New: func() any {
				const parserCacheCapacity = 5
				return make(parser.Cache, parserCacheCapacity)
			},
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	r.hasParsers = len(r.parsers) > 0
	r.hasDecoders = len(r.decoders) > 0
	r.hasFormatters = len(r.formatters) > 0

	formatters := make([]string, 0, len(r.formatters))
	reflectValueFormatters := make([]string, 0, len(r.formatters))
	for tag, f := range r.formatters {
		if _, ok := f.(ReflectValueFormatter); !ok {
			formatters = append(formatters, tag)

			continue
		}

		reflectValueFormatters = append(reflectValueFormatters, tag)
	}

	r.structureCache = cache.NewStructureCache(
		r.decoders.Tags(),
		maps.Keys(r.parsers),
		formatters,
		reflectValueFormatters,
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
		clear(parserCache)
		r.parserCachePool.Put(parserCache)
	}()

	fields := r.structureCache.Fields(t)

	for i := range fields {
		f := &fields[i]

		fieldValue := v.Field(f.Index)

		if r.skipFilled && !fieldValue.IsZero() {
			if r.hasFormatters && len(f.Formatters) > 0 {
				if err := r.applyFormatters(f, fieldValue); err != nil {
					return errors.WithMessagef(err, "format field `%s` in struct `%T`", f.Name, ptr)
				}
			}

			continue
		}

		var parsedSuccessfully bool
		if r.hasParsers {
			for _, parserName := range f.Parsers {
				p, ok := r.parsers[parserName]
				if !ok {
					continue
				}

				parsedValue, ok := p.Parse(req, f.StructField.Tag, parserCache)
				if !ok {
					continue
				}

				if err := value.Set(fieldValue, parsedValue); err != nil {
					return errors.Wrapf(err, "set `%s` value to field `%s` from tag `%s` for struct `%T`",
						parsedValue, f.Name, parserName, ptr)
				}

				parsedSuccessfully = true

				break
			}
		}

		if !parsedSuccessfully && f.HasDefault && fieldValue.IsZero() {
			if err := value.Set(fieldValue, f.DefaultValue); err != nil {
				return errors.Wrapf(err, "set default value for field `%s`", f.Name)
			}
		}

		if r.hasFormatters && len(f.Formatters) > 0 {
			if err := r.applyFormatters(f, fieldValue); err != nil {
				return errors.WithMessagef(err, "format field `%s` in struct `%T`", f.Name, ptr)
			}
		}
	}

	return nil
}

func (r *Roamer) applyFormatters(field *cache.Field, fieldValue reflect.Value) error {
	for _, name := range field.ReflectValueFormatters {
		f, ok := r.formatters[name].(ReflectValueFormatter)
		if !ok {
			continue
		}

		if err := f.FormatReflectValue(field.StructField.Tag, fieldValue); err != nil {
			return err
		}
	}

	ptr, ok := value.Pointer(fieldValue)
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
func (r *Roamer) parseBody(req *http.Request, ptr any) error {
	if !r.hasDecoders || req.ContentLength == 0 || req.Method == http.MethodGet {
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

	if err := d.Decode(req, ptr); err != nil {
		return errors.WithStack(rerr.DecodeError{
			Err: errors.WithMessagef(err, "decode `%s` request body for `%T`", contentType, ptr),
		})
	}

	return nil
}
