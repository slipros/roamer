package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPath(t *testing.T) {
	h := NewPath(nil)
	require.NotNil(t, h)
	require.Equal(t, TagPath, h.Tag())
}

func TestPath(t *testing.T) {
	pathParamValue := "1337"
	pathParam := "user_id"

	pathValueFunc := func(r *http.Request, name string) (string, bool) {
		_, after, found := strings.Cut(r.URL.Path, name+"/")
		if !found {
			return "", false
		}

		return after, true
	}

	type args struct {
		req           *http.Request
		pathValueFunc PathValueFunc
		tag           reflect.StructTag
	}
	tests := []struct {
		name string
		args func() args
		want any
	}{
		{
			name: "Get path variable from request",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", requestURL, pathParam, pathParamValue))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:           req,
					pathValueFunc: pathValueFunc,
					tag:           reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagPath, pathParam)),
				}
			},
			want: pathParamValue,
		},
		{
			name: "Get path variable from request - nil pathValueFunc",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", requestURL, pathParam, pathParamValue))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:           req,
					pathValueFunc: nil,
					tag:           reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagPath, pathParam)),
				}
			},
			want: "",
		},
		{
			name: "Get path variable from request - no path variable",
			args: func() args {
				rawURL, err := url.Parse(requestURL)
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:           req,
					pathValueFunc: pathValueFunc,
					tag:           reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagPath, pathParam)),
				}
			},
			want: "",
		},
		{
			name: "Get path variable from request - wrong tag path",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", requestURL, pathParam, pathParamValue))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:           req,
					pathValueFunc: pathValueFunc,
					tag:           reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagPath+"1", pathParam)),
				}
			},
			want: "",
		},
		{
			name: "Get path variable from request - wrong path param",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", requestURL, pathParam, pathParamValue))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:           req,
					pathValueFunc: pathValueFunc,
					tag:           reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagPath, pathParam+"1")),
				}
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			p := NewPath(args.pathValueFunc)
			value, exists := p.Parse(args.req, args.tag, nil)

			if tt.want == nil && exists {
				t.Errorf("Parse() want is nil, but value exists")
			}

			require.Equal(t, tt.want, value)
		})
	}
}

func TestServeMuxValueFromPath_Successfully(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *http.Request
		paramName   string
		expectedVal string
		expectedOk  bool
	}{
		{
			name: "value exists",
			setup: func() *http.Request {
				req := &http.Request{}
				req.SetPathValue("user_id", "123")
				return req
			},
			paramName:   "user_id",
			expectedVal: "123",
			expectedOk:  true,
		},
		{
			name: "value does not exist",
			setup: func() *http.Request {
				return &http.Request{}
			},
			paramName:   "user_id",
			expectedVal: "",
			expectedOk:  false,
		},
		{
			name: "value is empty",
			setup: func() *http.Request {
				req := &http.Request{}
				req.SetPathValue("user_id", "")
				return req
			},
			paramName:   "user_id",
			expectedVal: "",
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup()
			val, ok := ServeMuxValueFromPath(req, tt.paramName)
			require.Equal(t, tt.expectedVal, val)
			require.Equal(t, tt.expectedOk, ok)
		})
	}
}
