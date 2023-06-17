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
		_, after, found := strings.Cut(r.URL.Path, pathParam+"/")
		if !found {
			return "", false
		}

		before, _, found := strings.Cut(after, "/")
		if !found {
			return "", false
		}

		return before, true
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
