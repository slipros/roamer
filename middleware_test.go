package roamer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/mockroamer"
	"github.com/slipros/roamer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Field string `query:"field" json:"field"`
}

func TestMiddleware_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		setupRoamer   func() *Roamer
		validateCtx   func(t *testing.T, ctx context.Context)
		expectedCalls int
	}{
		{
			name: "Success parsing from query",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/?field=value", nil)
				return req
			},
			setupRoamer: func() *Roamer {
				return NewRoamer(
					WithParsers(parser.NewQuery()),
				)
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				var data testData
				err := ParsedDataFromContext(ctx, &data)
				require.NoError(t, err)
				assert.Equal(t, "value", data.Field)
			},
			expectedCalls: 1,
		},
		{
			name: "Success parsing from JSON body",
			setupRequest: func() *http.Request {
				body := strings.NewReader(`{"field": "json-value"}`)
				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", "application/json")
				req.ContentLength = int64(len(`{"field": "json-value"}`))
				return req
			},
			setupRoamer: func() *Roamer {
				mockDecoder := mockroamer.NewDecoder(t)
				mockDecoder.EXPECT().ContentType().Return("application/json").Maybe()
				mockDecoder.EXPECT().Tag().Return(decoder.TagJSON).Maybe()
				mockDecoder.EXPECT().
					Decode(mock.AnythingOfType("*http.Request"), mock.AnythingOfType("*roamer.testData")).
					Run(func(r *http.Request, ptr any) {
						dataPtr, ok := ptr.(*testData)
						if ok {
							dataPtr.Field = "json-value"
						}
					}).
					Return(nil)

				return NewRoamer(
					WithDecoders(mockDecoder),
				)
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				var data testData
				err := ParsedDataFromContext(ctx, &data)
				require.NoError(t, err)
				assert.Equal(t, "json-value", data.Field)
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.validateCtx(t, r.Context())
			})

			middleware := Middleware[testData](tt.setupRoamer())
			handler := middleware(next)

			req := tt.setupRequest()
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCalls, callCount, "next handler was not called expected number of times")
		})
	}
}

func TestMiddleware_Failure(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		setupRoamer   func() *Roamer
		validateCtx   func(t *testing.T, ctx context.Context)
		expectedCalls int
	}{
		{
			name: "Parsing error",
			setupRequest: func() *http.Request {
				body := strings.NewReader(`{"field": "test"}`)
				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", "application/json")
				req.ContentLength = int64(len(`{"field": "test"}`))
				return req
			},
			setupRoamer: func() *Roamer {
				mockDecoder := mockroamer.NewDecoder(t)
				mockDecoder.EXPECT().ContentType().Return("application/json").Maybe()
				mockDecoder.EXPECT().Tag().Return(decoder.TagJSON).Maybe()
				mockDecoder.EXPECT().
					Decode(mock.AnythingOfType("*http.Request"), mock.AnythingOfType("*roamer.testData")).
					Return(errBigBad)

				return NewRoamer(
					WithDecoders(mockDecoder),
				)
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				var data testData
				err := ParsedDataFromContext(ctx, &data)
				require.Error(t, err)
				assert.True(t, errors.Is(err, errBigBad))
			},
			expectedCalls: 1,
		},
		{
			name: "Nil roamer",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				return req
			},
			setupRoamer: func() *Roamer {
				return nil
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				// Context should remain unchanged
				_, ok := ctx.Value(ContextKeyParsedData).(*testData)
				assert.False(t, ok)
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.validateCtx(t, r.Context())
			})

			middleware := Middleware[testData](tt.setupRoamer())
			handler := middleware(next)

			req := tt.setupRequest()
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCalls, callCount, "next handler was not called expected number of times")
		})
	}
}

func TestSliceMiddleware_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		setupRoamer   func() *Roamer
		validateCtx   func(t *testing.T, ctx context.Context)
		expectedCalls int
	}{
		{
			name: "Success parsing slice",
			setupRequest: func() *http.Request {
				body := strings.NewReader(`["value1", "value2"]`)
				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", decoder.ContentTypeJSON)
				req.ContentLength = int64(len(`["value1", "value2"]`))
				return req
			},
			setupRoamer: func() *Roamer {
				mockDecoder := mockroamer.NewDecoder(t)
				mockDecoder.EXPECT().ContentType().Return(decoder.ContentTypeJSON).Maybe()
				mockDecoder.EXPECT().Tag().Return(decoder.TagJSON).Maybe()

				mockDecoder.EXPECT().
					Decode(mock.AnythingOfType("*http.Request"), mock.AnythingOfType("*[]string")).
					Run(func(r *http.Request, ptr any) {
						slicePtr, ok := ptr.(*[]string)
						if ok {
							*slicePtr = []string{"value1", "value2"}
						}
					}).
					Return(nil)

				return NewRoamer(
					WithDecoders(mockDecoder),
				)
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				var data []string
				err := ParsedDataFromContext(ctx, &data)
				require.NoError(t, err)
				assert.Equal(t, []string{"value1", "value2"}, data)
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.validateCtx(t, r.Context())
			})

			middleware := SliceMiddleware[string](tt.setupRoamer())
			handler := middleware(next)

			req := tt.setupRequest()
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCalls, callCount, "next handler was not called expected number of times")
		})
	}
}

func TestSliceMiddleware_Failure(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		setupRoamer   func() *Roamer
		validateCtx   func(t *testing.T, ctx context.Context)
		expectedCalls int
	}{
		{
			name: "Parsing error in slice",
			setupRequest: func() *http.Request {
				body := strings.NewReader(`["value1", "value2"]`)
				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", decoder.ContentTypeJSON)
				req.ContentLength = int64(len(`["value1", "value2"]`))
				return req
			},
			setupRoamer: func() *Roamer {
				mockDecoder := mockroamer.NewDecoder(t)
				mockDecoder.EXPECT().ContentType().Return(decoder.ContentTypeJSON).Maybe()
				mockDecoder.EXPECT().Tag().Return(decoder.TagJSON).Maybe()

				mockDecoder.EXPECT().
					Decode(mock.AnythingOfType("*http.Request"), mock.AnythingOfType("*[]string")).
					Return(errBigBad)

				return NewRoamer(
					WithDecoders(mockDecoder),
				)
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				var data []string
				err := ParsedDataFromContext(ctx, &data)
				require.Error(t, err)
				assert.True(t, errors.Is(err, errBigBad))
			},
			expectedCalls: 1,
		},
		{
			name: "Nil roamer for slice",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				return req
			},
			setupRoamer: func() *Roamer {
				return nil
			},
			validateCtx: func(t *testing.T, ctx context.Context) {
				// Context should remain unchanged
				_, ok := ctx.Value(ContextKeyParsedData).(*[]string)
				assert.False(t, ok)
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.validateCtx(t, r.Context())
			})

			middleware := SliceMiddleware[string](tt.setupRoamer())
			handler := middleware(next)

			req := tt.setupRequest()
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCalls, callCount, "next handler was not called expected number of times")
		})
	}
}
