package parser

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	type args struct {
		query url.Values
		tag   reflect.StructTag
	}
	tests := []struct {
		name       string
		args       args
		value      any
		exists     bool
		emptyCache bool
	}{
		{
			name: "Exists",
			args: args{
				query: url.Values{
					"agent": []string{"test"},
				},
				tag: reflect.StructTag(`query:"agent"`),
			},
			value:  "test agent",
			exists: true,
		},
		{
			name: "NotExists",
			args: args{
				tag: reflect.StructTag(`query:"agent"`),
			},
			exists: false,
		},
		{
			name: "NoTag",
			args: args{
				query: url.Values{
					"agent": []string{"test"},
				},
			},
			exists:     false,
			emptyCache: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := http.Request{URL: &url.URL{RawQuery: tt.args.query.Encode()}}
			cache := make(Cache)

			parse := NewQuery(",")
			tagName, value, exists := parse(&req, tt.args.tag, cache)
			if exists && tagName != TagQuery {
				t.Errorf("Query() tag name = %v, want %v", tagName, TagQuery)
			}
			if exists != tt.exists {
				t.Errorf("Query() exists = %v, want %v", exists, tt.exists)
			}
			if !reflect.DeepEqual(value, value) {
				t.Errorf("Query() value = %v, want %v", value, tt.value)
			}

			if !tt.emptyCache {
				_, ok := cache[cacheKeyQuery]
				require.True(t, ok, "Query() empty cache")
			}
		})
	}
}
