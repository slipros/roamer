package parser

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeader(t *testing.T) {
	type args struct {
		header http.Header
		tag    reflect.StructTag
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
				header: http.Header{
					"User-Agent": []string{"test agent"},
				},
				tag: reflect.StructTag(`header:"User-Agent"`),
			},
			value:      "test agent",
			exists:     true,
			emptyCache: true,
		},
		{
			name: "NotExists",
			args: args{
				tag: reflect.StructTag(`header:"User-Agent"`),
			},
			exists:     false,
			emptyCache: true,
		},
		{
			name: "NoTag",
			args: args{
				header: http.Header{
					"User-Agent": []string{"test agent"},
				},
			},
			exists:     false,
			emptyCache: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := http.Request{Header: tt.args.header}
			cache := make(Cache)

			tagName, value, exists := Header(&req, tt.args.tag, cache)
			if exists && tagName != TagHeader {
				t.Errorf("Header() tag name = %v, want %v", tagName, TagHeader)
			}
			if exists != tt.exists {
				t.Errorf("Header() exists = %v, want %v", exists, tt.exists)
			}
			if !reflect.DeepEqual(value, value) {
				t.Errorf("Header() value = %v, want %v", value, tt.value)
			}

			if tt.emptyCache {
				require.Empty(t, cache, "Header() not empty cache %v", cache)
			}
		})
	}
}
