package roamer

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipBodyStruct implements BodyDecodeSkipper. The skip flag is set by the test
// before Parse, mirroring the intended usage where the decision is made by the
// caller prior to parsing.
type skipBodyStruct struct {
	skip bool

	BodyValue  string `json:"body_value"`
	QueryValue string `query:"query_value"`
}

func (s *skipBodyStruct) SkipBodyDecode() bool {
	return s.skip
}

// TestRoamer_Parse_BodyDecodeSkipper_Successfully verifies that body decoding is
// skipped when SkipBodyDecode returns true, while other request parts are still
// parsed, and that decoding proceeds normally when it returns false.
func TestRoamer_Parse_BodyDecodeSkipper_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		skip          bool
		body          string
		expectedBody  string
		expectedQuery string
	}{
		{
			name:          "skip decoding, query still parsed",
			skip:          true,
			body:          `{"body_value":"fromBody"}`,
			expectedBody:  "",
			expectedQuery: "fromQuery",
		},
		{
			name:          "decode body when not skipped",
			skip:          false,
			body:          `{"body_value":"fromBody"}`,
			expectedBody:  "fromBody",
			expectedQuery: "fromQuery",
		},
		{
			name:          "skip decoding ignores malformed body",
			skip:          true,
			body:          `{invalid`,
			expectedBody:  "",
			expectedQuery: "fromQuery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			req, err := http.NewRequest(http.MethodPost,
				"http://example.com?query_value=fromQuery", bytes.NewReader([]byte(tt.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", decoder.ContentTypeJSON)

			r := NewRoamer(
				WithDecoders(decoder.NewJSON()),
				WithParsers(parser.NewQuery()),
			)

			target := skipBodyStruct{skip: tt.skip}

			// act
			err = r.Parse(req, &target)

			// assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, target.BodyValue, "body field")
			assert.Equal(t, tt.expectedQuery, target.QueryValue, "query field")
		})
	}
}

// TestRoamer_Parse_BodyDecodeSkipper_Failure verifies that a malformed body
// still produces a decode error when SkipBodyDecode returns false, confirming
// that skipping is the only thing the interface suppresses.
func TestRoamer_Parse_BodyDecodeSkipper_Failure(t *testing.T) {
	// arrange
	req, err := http.NewRequest(http.MethodPost,
		"http://example.com", bytes.NewReader([]byte(`{invalid`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", decoder.ContentTypeJSON)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))

	target := skipBodyStruct{skip: false}

	// act
	err = r.Parse(req, &target)

	// assert
	require.Error(t, err)
	_, ok := IsDecodeError(err)
	assert.True(t, ok, "expected a decode error")
}

// TestRoamer_Parse_BodyDecodeSkipper_PreservesBody_Successfully verifies the
// interaction between WithPreserveBody and BodyDecodeSkipper: when skipped the
// body is left intact (and still readable once downstream), and when not skipped
// the normal preserve-body path keeps the body readable.
func TestRoamer_Parse_BodyDecodeSkipper_PreservesBody_Successfully(t *testing.T) {
	const body = `{"body_value":"fromBody"}`

	tests := []struct {
		name         string
		skip         bool
		expectedBody string
	}{
		{
			name:         "skip leaves body untouched and readable",
			skip:         true,
			expectedBody: "",
		},
		{
			name:         "no skip decodes and preserves body",
			skip:         false,
			expectedBody: "fromBody",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			req, err := http.NewRequest(http.MethodPost,
				"http://example.com", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", decoder.ContentTypeJSON)

			r := NewRoamer(
				WithDecoders(decoder.NewJSON()),
				WithPreserveBody(),
			)

			target := skipBodyStruct{skip: tt.skip}

			// act
			err = r.Parse(req, &target)

			// assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, target.BodyValue, "decoded body field")

			remaining, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Equal(t, body, string(remaining), "body must remain readable downstream")
		})
	}
}
