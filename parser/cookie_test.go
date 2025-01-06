package parser

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCookie(t *testing.T) {
	h := NewCookie()
	require.NotNil(t, h)
	require.Equal(t, TagCookie, h.Tag())
}

func TestCookie(t *testing.T) {
	cookie := "ref"
	cookieValue := "test"

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
			name: "Get cookie value from request cookie",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.AddCookie(&http.Cookie{
					Name:  cookie,
					Value: cookieValue,
				})

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s""`, TagCookie, cookie)),
				}
			},
			want: &http.Cookie{
				Name:  cookie,
				Value: cookieValue,
			},
		},
		{
			name: "Get cookie value from request cookie - empty struct tag",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.AddCookie(&http.Cookie{
					Name:  cookie,
					Value: cookieValue,
				})

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

			h := NewCookie()
			value, exists := h.Parse(args.req, args.tag, nil)

			if tt.want == nil && exists {
				t.Errorf("Parse() want is nil, but value exists")
			}

			require.Equal(t, tt.want, value)
		})
	}
}
