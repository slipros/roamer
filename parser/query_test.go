package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewQuery(t *testing.T) {
	q := NewQuery(",")
	require.NotNil(t, q)
	require.Equal(t, TagQuery, q.Tag())
}

func TestQuery(t *testing.T) {
	queryName := "user_id"
	queryValue := "1337"

	type args struct {
		req   *http.Request
		tag   reflect.StructTag
		cache Cache
	}
	tests := []struct {
		name string
		args func() args
		want any
	}{
		{
			name: "get value from query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(map[string]any),
				}
			},
			want: queryValue,
		},
		{
			name: "get value from cached query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				cache := make(map[string]any, 1)
				cache[cacheKeyQuery] = q

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: cache,
				}
			},
			want: queryValue,
		},
		{
			name: "get value from query - no query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(map[string]any),
				}
			},
			want: "",
		},
		{
			name: "get value from query - wrong query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName+"1", queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(map[string]any),
				}
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			q := NewQuery(",")

			value, exists := q.Parse(args.req, args.tag, args.cache)

			if tt.want == nil && exists {
				t.Errorf("Parse() does not want want, but it is exists")
			}

			require.Equal(t, tt.want, value)
		})
	}
}
