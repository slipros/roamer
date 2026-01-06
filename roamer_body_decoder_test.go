package roamer

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/slipros/roamer/decoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bodyTestStruct struct {
	BodyBytes  []byte    `decoder:"body"`
	BodyReader io.Reader `decoder:"body"`
	Other      string    `json:"other"`
}

type bodyBytesOnly struct {
	Body []byte `decoder:"body"`
}

type bodyReaderOnly struct {
	Body io.Reader `decoder:"body"`
}

func TestRoamer_Parse_BodyDecoder_Successfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		body          string
		setup         func() *Roamer
		targetFactory func() any
		verify        func(*testing.T, any, string)
	}{
		{
			name: "capture body as bytes",
			body: `{"other":"value"}`,
			setup: func() *Roamer {
				return NewRoamer(WithDecoders(decoder.NewJSON()))
			},
			targetFactory: func() any { return &bodyBytesOnly{} },
			verify: func(t *testing.T, res any, expectedBody string) {
				target := res.(*bodyBytesOnly)
				assert.Equal(t, expectedBody, string(target.Body))
			},
		},
		{
			name: "capture body as reader",
			body: `{"other":"value"}`,
			setup: func() *Roamer {
				return NewRoamer(WithDecoders(decoder.NewJSON()))
			},
			targetFactory: func() any { return &bodyReaderOnly{} },
			verify: func(t *testing.T, res any, expectedBody string) {
				target := res.(*bodyReaderOnly)
				require.NotNil(t, target.Body)
				content, err := io.ReadAll(target.Body)
				require.NoError(t, err)
				assert.Equal(t, expectedBody, string(content))
			},
		},
		{
			name: "capture body as both bytes and reader mixed with json decoder",
			body: `{"other":"value"}`,
			setup: func() *Roamer {
				return NewRoamer(WithDecoders(decoder.NewJSON()))
			},
			targetFactory: func() any { return &bodyTestStruct{} },
			verify: func(t *testing.T, res any, expectedBody string) {
				target := res.(*bodyTestStruct)
				assert.Equal(t, "value", target.Other)
				assert.Equal(t, expectedBody, string(target.BodyBytes))

				require.NotNil(t, target.BodyReader)
				content, err := io.ReadAll(target.BodyReader)
				require.NoError(t, err)
				assert.Equal(t, expectedBody, string(content))
			},
		},
		{
			name: "capture body with preserve body enabled",
			body: `{"other":"value"}`,
			setup: func() *Roamer {
				return NewRoamer(
					WithDecoders(decoder.NewJSON()),
					WithPreserveBody(),
				)
			},
			targetFactory: func() any { return &bodyTestStruct{} },
			verify: func(t *testing.T, res any, expectedBody string) {
				target := res.(*bodyTestStruct)
				assert.Equal(t, "value", target.Other)
				assert.Equal(t, expectedBody, string(target.BodyBytes))
			},
		},
		{
			name: "works without any registered decoders",
			body: `some raw data`,
			setup: func() *Roamer {
				return NewRoamer()
			},
			targetFactory: func() any { return &bodyBytesOnly{} },
			verify: func(t *testing.T, res any, expectedBody string) {
				target := res.(*bodyBytesOnly)
				assert.Equal(t, expectedBody, string(target.Body))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r := tt.setup()
			target := tt.targetFactory()

			err := r.Parse(req, target)
			require.NoError(t, err)

			tt.verify(t, target, tt.body)

			// Verify req.Body state based on configuration
			// If we are preserving body, we should be able to read it again
			// Note: We need to check if the specific test setup enabled preserve body
			// But since we can't easily inspect the roamer instance private fields here,
			// we'll rely on the functional verification.
		})
	}
}

func TestRoamer_Parse_BodyDecoder_Failure(t *testing.T) {
	t.Parallel()

	// Currently, body reading is quite robust, but we can test edge cases
	// typically covered by io errors, but mocking io.Reader in http.Request is tricky
	// without custom request creation.
	// For now, we will add failure cases if we identify strict requirements that can fail.

	tests := []struct {
		name          string
		setup         func() *Roamer
		targetFactory func() any
		reqFactory    func() *http.Request
		errorContains string
	}{
		{
			name: "unsupported type for body tag",
			setup: func() *Roamer {
				return NewRoamer()
			},
			targetFactory: func() any {
				return &struct {
					Body int `decoder:"body"`
				}{}
			},
			reqFactory: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader("123"))
				return req
			},
			errorContains: "set `body` value to field `Body`", // Expecting assignment error
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := tt.reqFactory()
			r := tt.setup()
			target := tt.targetFactory()

			err := r.Parse(req, target)
			require.Error(t, err)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}
		})
	}
}
