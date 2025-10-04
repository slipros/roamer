package parser

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const requestURL = "test.com"

func TestNewHeader(t *testing.T) {
	h := NewHeader()
	require.NotNil(t, h)
	assert.Equal(t, TagHeader, h.Tag())
}

func TestHeader(t *testing.T) {
	header := "User-Agent"
	headerValue := "test"

	type args struct {
		req *http.Request
		tag reflect.StructTag
	}
	tests := []struct {
		name string
		args func() args
		want any
	}{
		{
			name: "Get header value from request header",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add(header, headerValue)

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s""`, TagHeader, header)),
				}
			},
			want: headerValue,
		},
		{
			name: "Get header value from first of few request headers",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add("Referer", "referer")
				req.Header.Add("X-Referer", "x-referer")

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s""`, TagHeader, "Referer,X-Referer")),
				}
			},
			want: "referer",
		},
		{
			name: "Get header value from second of few request headers",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add(header, headerValue)
				req.Header.Add("X-Referer", "x-referer")

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s""`, TagHeader, "Referer,X-Referer")),
				}
			},
			want: "x-referer",
		},
		{
			name: "Get header value from request header - empty request header",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s""`, TagHeader, header)),
				}
			},
			want: "",
		},
		{
			name: "Get header value from request header - empty struct tag",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add(header, headerValue)

				return args{
					req: req,
					tag: "",
				}
			},
			want: "",
		},
		{
			name: "Get header value with whitespace in tag list",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add("X-Forwarded-For", "192.168.1.1")

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, "X-Real-IP, X-Forwarded-For")),
				}
			},
			want: "192.168.1.1",
		},
		{
			name: "Get header value when none of fallback headers exist",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, "X-Real-IP,X-Forwarded-For,X-Client-IP")),
				}
			},
			want: "",
		},
		{
			name: "Get header value with empty string in comma-separated list",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add("X-Custom", "value")

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, ",X-Custom,")),
				}
			},
			want: "value",
		},
		{
			name: "Get header with case-insensitive matching",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.Header.Add("content-type", "application/json")

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, "Content-Type")),
				}
			},
			want: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := tt.args()

			h := NewHeader()
			value, exists := h.Parse(args.req, args.tag, nil)

			if tt.want == nil {
				assert.False(t, exists, "Parse() want is nil, but value exists")
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

// TestHeader_EdgeCases tests edge cases and boundary conditions
func TestHeader_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func() (*http.Request, reflect.StructTag)
		wantValue  any
		wantExists bool
	}{
		{
			name: "header with multiple values returns first",
			setup: func() (*http.Request, reflect.StructTag) {
				req, _ := http.NewRequest(http.MethodGet, requestURL, nil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Accept", "text/html")
				return req, reflect.StructTag(`header:"Accept"`)
			},
			wantValue:  "application/json",
			wantExists: true,
		},
		{
			name: "empty header value should return false",
			setup: func() (*http.Request, reflect.StructTag) {
				req, _ := http.NewRequest(http.MethodGet, requestURL, nil)
				req.Header.Set("X-Empty", "")
				return req, reflect.StructTag(`header:"X-Empty"`)
			},
			wantValue:  "",
			wantExists: false,
		},
		{
			name: "header with only commas in tag",
			setup: func() (*http.Request, reflect.StructTag) {
				req, _ := http.NewRequest(http.MethodGet, requestURL, nil)
				req.Header.Set("X-Test", "value")
				return req, reflect.StructTag(`header:",,,"`)
			},
			wantValue:  "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, tag := tt.setup()
			h := NewHeader()

			// Should not panic
			value, exists := h.Parse(req, tag, nil)

			assert.Equal(t, tt.wantExists, exists)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}
