package roamer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
)

var errBigBad = errors.New("big bad error")

func TestRoamer_Parse(t *testing.T) {
	type fields struct {
		skipFilled bool
		decoders   Decoders
		parsers    Parsers
	}
	type args struct {
		req *http.Request
		ptr any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Roamer{
				skipFilled: tt.fields.skipFilled,
				decoders:   tt.fields.decoders,
				parsers:    tt.fields.parsers,
			}
			if err := r.Parse(tt.args.req, tt.args.ptr); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkParse_With_Body_Header_Query(b *testing.B) {
	toJSON := func(v any) (int, io.Reader, error) {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(&v); err != nil {
			return 0, nil, err
		}

		return buffer.Len(), &buffer, nil
	}

	query := make(url.Values)
	query.Add("int", "9223372036854775807")
	query.Add("int8", "127")
	query.Add("int16", "32767")
	query.Add("int32", "2147483647")
	query.Add("int64", "9223372036854775807")
	query.Add("time", "2002-10-02T15:00:00.05Z")
	query.Add("url", "http://google.com")

	header := make(http.Header)
	header.Add("User-Agent", "agent 1337")

	bodyLen, body, err := toJSON(
		struct {
			Strings []string `json:"strings"`
		}{
			Strings: []string{"1", "2"},
		},
	)
	if err != nil {
		b.Fatal(err)
	}

	header.Add("Content-Type", decoder.ContentTypeJSON)
	header.Add("Content-Length", strconv.Itoa(bodyLen))

	type Data struct {
		Int   int       `query:"int"`
		Int8  int8      `query:"int8"`
		Int32 int32     `query:"int32"`
		Int64 int64     `query:"int64"`
		Time  time.Time `query:"time"`
		Url   url.URL   `query:"url"`

		UserAgent string   `header:"User-Agent"`
		Strings   []string `json:"strings"`
	}

	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			RawQuery: query.Encode(),
		},
		Header:        header,
		ContentLength: int64(bodyLen),
		Body:          io.NopCloser(body),
	}

	r := NewRoamer(
		WithSkipFilled(false),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
	)

	var d Data

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(&req, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_With_Body_Header_Query_FastStructFieldParser(b *testing.B) {
	toJSON := func(v any) (int, io.Reader, error) {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(&v); err != nil {
			return 0, nil, err
		}

		return buffer.Len(), &buffer, nil
	}

	query := make(url.Values, 7)
	query.Add("int", "9223372036854775807")
	query.Add("int8", "127")
	query.Add("int16", "32767")
	query.Add("int32", "2147483647")
	query.Add("int64", "9223372036854775807")
	query.Add("time", "2002-10-02T15:00:00.05Z")
	query.Add("url", "http://google.com")

	header := make(http.Header)
	header.Add("User-Agent", "agent 1337")

	bodyLen, body, err := toJSON(
		struct {
			Strings []string `json:"strings"`
		}{
			Strings: []string{"1", "2"},
		},
	)
	if err != nil {
		b.Fatal(err)
	}

	header.Add("Content-Type", decoder.ContentTypeJSON)
	header.Add("Content-Length", strconv.Itoa(bodyLen))

	type Data struct {
		Int   int       `query:"int"`
		Int8  int8      `query:"int8"`
		Int32 int32     `query:"int32"`
		Int64 int64     `query:"int64"`
		Time  time.Time `query:"time"`
		Url   url.URL   `query:"url"`

		UserAgent string   `header:"User-Agent"`
		Strings   []string `json:"strings"`
	}

	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			RawQuery: query.Encode(),
		},
		Header:        header,
		ContentLength: int64(bodyLen),
		Body:          io.NopCloser(body),
	}

	r := NewRoamer(
		WithSkipFilled(false),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
		WithExperimentalFastStructFieldParser(),
	)

	var d Data

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(&req, &d); err != nil {
			b.Fatal(err)
		}
	}
}
