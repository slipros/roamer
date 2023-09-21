// Package roamer provides flexible http request parser.
package roamer

import (
	"context"
	"net/http"
	"reflect"

	"github.com/pkg/errors"

	roamerError "github.com/SLIpros/roamer/err"
	"github.com/SLIpros/roamer/parser"
	"github.com/SLIpros/roamer/value"
)

// AfterParser will be called after http request parsing.
//
//go:generate mockery --name=AfterParser --outpkg=mock --output=./mock
type AfterParser interface {
	AfterParse(ctx context.Context) error
}

// Roamer flexible http request parser.
type Roamer struct {
	parsers    Parsers
	decoders   Decoders
	skipFilled bool
}

// NewRoamer creates and returns new roamer.
func NewRoamer(opts ...OptionsFunc) *Roamer {
	r := Roamer{
		parsers:    make(Parsers),
		decoders:   make(Decoders),
		skipFilled: true,
	}

	for _, opt := range opts {
		opt(&r)
	}

	return &r
}

// Parse parses http request into ptr.
//
// ptr can implement AfterParser to execute some logic after parsing.
func (r *Roamer) Parse(req *http.Request, ptr any) error {
	if ptr == nil {
		return errors.WithMessage(roamerError.NilValue, "ptr")
	}

	t := reflect.TypeOf(ptr)
	if t.Kind() != reflect.Pointer {
		return errors.WithMessagef(roamerError.NotPtr, "`%T`", ptr)
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
		return errors.WithMessagef(roamerError.NotSupported, "`%T`", ptr)
	}

	if p, ok := ptr.(AfterParser); ok {
		return p.AfterParse(req.Context())
	}

	return nil
}

// parseStruct parses structure from http request into a ptr.
func (r *Roamer) parseStruct(req *http.Request, ptr any) error {
	if err := r.parseBody(req, ptr); err != nil {
		return err
	}

	if len(r.parsers) == 0 {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(ptr))
	t := v.Type()

	cache := make(parser.Cache)
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		if !fieldType.IsExported() || len(fieldType.Tag) == 0 {
			continue
		}

		fieldValue := v.Field(i)
		if r.skipFilled && !fieldValue.IsZero() {
			continue
		}

		for tag, p := range r.parsers {
			parsedValue, ok := p.Parse(req, fieldType.Tag, cache)
			if !ok {
				continue
			}

			if err := value.Set(&fieldValue, parsedValue); err != nil {
				return errors.WithMessagef(err, "set `%s` value to field `%s` from tag `%s` for struct `%T`",
					parsedValue, fieldType.Name, tag, ptr)
			}
		}
	}

	return nil
}

// parseStruct parses body from http request into a ptr.
func (r *Roamer) parseBody(req *http.Request, ptr any) error {
	if req.Method == http.MethodGet || req.ContentLength == 0 {
		return nil
	}

	contentType := req.Header.Get("Content-Type")
	d, ok := r.decoders[contentType]
	if !ok {
		return nil
	}

	if err := d.Decode(req, ptr); err != nil {
		return &roamerError.DecodeError{
			Err: errors.WithMessagef(err, "decode `%s` request body in `%T`", contentType, ptr),
		}
	}

	return nil
}
