package decoder

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewXML(t *testing.T) {
	x := NewXML()
	require.NotNil(t, x)
	require.Equal(t, ContentTypeXML, x.ContentType())

	x = NewXML(WithContentType[*XML]("test"))
	require.NotNil(t, x)
	require.Equal(t, "test", x.ContentType())
}

func TestXML_Decode(t *testing.T) {
	type args struct {
		req  *http.Request
		ptr  any
		want any
	}
	tests := []struct {
		name    string
		args    func() args
		wantErr bool
	}{
		{
			name: "Fill struct",
			args: func() args {
				type Data struct {
					Field1 string `xml:"field_1"`
					Field2 int    `xml:"field_2"`
				}

				data := Data{
					Field1: "field1",
					Field2: 2,
				}

				body := toXML(t, &data)
				req, err := http.NewRequest(http.MethodPost, requestURL, body)
				require.NoError(t, err)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &data,
				}
			},
		},
		{
			name: "Error request body is nil",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				return args{
					req: req,
				}
			},
			wantErr: true,
		},
		{
			name: "Error invalid request body",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader("<></>"))
				require.NoError(t, err)

				return args{
					req: req,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewXML()
			args := tt.args()

			if err := x.Decode(args.req, args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			require.Equal(t, args.want, args.ptr)
		})
	}
}
