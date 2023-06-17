package parser

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

const requestURL = "test.com"

func TestNewHeader(t *testing.T) {
	h := NewHeader()
	require.NotNil(t, h)
	require.Equal(t, TagHeader, h.Tag())
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
			name: "get want from request header",
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
			name: "empty request header",
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
			name: "empty struct tag",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			h := NewHeader()
			value, exists := h.Parse(args.req, args.tag, nil)

			if tt.want == nil && exists {
				t.Errorf("Parse() does not want want, but it is exists")
			}

			require.Equal(t, tt.want, value)
		})
	}
}
