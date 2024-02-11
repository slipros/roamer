package decoder

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewJSON(t *testing.T) {
	j := NewJSON()
	require.NotNil(t, j)
	require.Equal(t, ContentTypeJSON, j.ContentType())

	j = NewJSON(WithContentType[*JSON]("test"))
	require.NotNil(t, j)
	require.Equal(t, "test", j.ContentType())
}

func TestJSON_Decode(t *testing.T) {
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
			name: "Success fill struct",
			args: func() args {
				type Data struct {
					Field1 string `json:"field_1"`
					Field2 int    `json:"field_2"`
				}

				data := Data{
					Field1: "field1",
					Field2: 2,
				}

				body := toJSON(t, &data)
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
			name: "Fill slice of strings",
			args: func() args {
				data := []string{"1", "2"}

				body := toJSON(t, &data)
				req, err := http.NewRequest(http.MethodPost, requestURL, body)
				require.NoError(t, err)

				return args{
					req:  req,
					ptr:  &[]string{},
					want: &data,
				}
			},
		},
		{
			name: "Error nil request body",
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
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader("{]"))
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
			j := NewJSON()
			args := tt.args()

			if err := j.Decode(args.req, args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			require.Equal(t, args.want, args.ptr)
		})
	}
}
