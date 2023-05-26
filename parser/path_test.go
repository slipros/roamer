package parser

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	type args struct {
		valueFunc PathValueFunc
		tag       reflect.StructTag
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
				valueFunc: func(name string, r *http.Request) (string, bool) {
					return "1337", true
				},
				tag: reflect.StructTag(`path:"user_id"`),
			},
			value:      "1337",
			exists:     true,
			emptyCache: true,
		},
		{
			name: "NotExists",
			args: args{
				valueFunc: func(name string, r *http.Request) (string, bool) {
					return "", false
				},
				tag: reflect.StructTag(`path:"user_id"`),
			},
			exists:     false,
			emptyCache: true,
		},
		{
			name: "NoTag",
			args: args{
				valueFunc: func(name string, r *http.Request) (string, bool) {
					return "", false
				},
			},
			exists:     false,
			emptyCache: true,
		},
		{
			name: "NilValueFunc",
			args: args{
				tag: reflect.StructTag(`path:"user_id"`),
			},
			exists:     false,
			emptyCache: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := http.Request{}
			cache := make(Cache)

			path := NewPath(tt.args.valueFunc)
			tagName, value, exists := path(&req, tt.args.tag, cache)
			if exists && tagName != TagPath {
				t.Errorf("Path() tag name = %v, want %v", tagName, TagPath)
			}
			if exists != tt.exists {
				t.Errorf("Path() exists = %v, want %v", exists, tt.exists)
			}
			if !reflect.DeepEqual(value, value) {
				t.Errorf("Path() value = %v, want %v", value, tt.value)
			}

			if tt.emptyCache {
				require.Empty(t, cache, "Path() not empty cache %v", cache)
			}
		})
	}
}
