// Package roamer flexible http request parser.
package roamer

import (
	"context"
	"net/http"
	"reflect"

	"github.com/pkg/errors"

	"github.com/SLIpros/roamer/decoder"
	roamerError "github.com/SLIpros/roamer/error"
	"github.com/SLIpros/roamer/parser"
	"github.com/SLIpros/roamer/value"
)

// TODO: Сохранять остаток json
// TODO: Попробовать стандартный router
// TODO: Поддержка файлов

const (
	SplitSymbol = ","
)

type Prepare interface {
	Prepare(ctx context.Context) error
}

// Roamer flexible http request parser.
type Roamer struct {
	skipFilled bool
	decoders   decoder.Decoders
	parsers    parser.Parsers
}

// NewRoamer creates and returns new roamer.
func NewRoamer(opts ...OptionsFunc) *Roamer {
	r := Roamer{
		skipFilled: true,
		decoders: decoder.Decoders{
			decoder.ContentTypeJSON:           decoder.NewJSON(),
			decoder.ContentTypeXML:            decoder.NewXML(),
			decoder.ContentTypeFormURLEncoded: decoder.NewFormURLEncoded(SplitSymbol),
		},
		parsers: parser.Parsers{
			parser.TagQuery: parser.NewQuery(SplitSymbol),
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	return &r
}

// Parse parse http request to ptr.
func (r *Roamer) Parse(req *http.Request, ptr any) error {
	if ptr == nil {
		return errors.WithMessage(roamerError.ErrNil, "ptr")
	}

	t := reflect.TypeOf(ptr)
	if t.Kind() != reflect.Pointer {
		return errors.WithMessagef(roamerError.ErrNotPtr, "`%T`", ptr)
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
		return errors.WithMessagef(roamerError.ErrNotSupported, "`%T`", ptr)
	}

	if prepare, ok := ptr.(Prepare); ok {
		return prepare.Prepare(req.Context())
	}

	return nil
}

// parseStruct parser structure from http request to pointer.
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

// parseStruct parser body from http request to pointer.
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
		return errors.WithMessagef(err, "decode `%s` request body in `%T`",
			contentType, ptr)
	}

	return nil
}
