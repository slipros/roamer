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
	q := NewQuery()
	require.NotNil(t, q)
	require.Equal(t, TagQuery, q.Tag())

	q = NewQuery(WithDisabledSplit())
	require.NotNil(t, q)
	require.False(t, q.split)

	q = NewQuery(WithSplitSymbol(";"))
	require.NotNil(t, q)
	require.Equal(t, ";", q.splitSymbol)
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
		name      string
		args      func() args
		want      any
		notExists bool
	}{
		{
			name: "Get value from query",
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
					cache: make(Cache),
				}
			},
			want: queryValue,
		},
		{
			name: "Get value from cached query",
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
			name: "Get value from query - no query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: "",
		},
		{
			name: "Get value from query - wrong query",
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
					cache: make(Cache),
				}
			},
			want: "",
		},
		{
			name: "Get value from array query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)
				q.Add(queryName, queryValue+"2")

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: []string{queryValue, queryValue + "2"},
		},
		{
			name: "Get value from query with split symbol",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue+","+queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: []string{queryValue, queryValue},
		},
		{
			name:      "Wrong tag",
			notExists: true,
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue+","+queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, queryName)),
					cache: make(Cache),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			q := NewQuery()

			value, exists := q.Parse(args.req, args.tag, args.cache)
			if tt.notExists && exists {
				t.Errorf("Parse() want not exists, but value exists")
			}

			if tt.want == nil && exists {
				t.Errorf("Parse() want is nil, but value exists")
			}

			if !tt.notExists {
				require.Equal(t, tt.want, value)
			}
		})
	}
}
