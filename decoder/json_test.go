package decoder

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewJSON(t *testing.T) {
	d := NewJSON()
	require.NotNil(t, d)
	require.Equal(t, ContentTypeJSON, d.ContentType())
}

func TestJSON_Decode(t *testing.T) {
	type args struct {
		body io.Reader
		ptr  any
		want any
	}
	tests := []struct {
		name    string
		args    func() args
		wantErr bool
	}{
		{
			name: "SliceOfStrings",
			args: func() args {
				data := []string{"1", "2"}

				return args{
					body: toJSON(t, &data),
					ptr:  &[]string{},
					want: &data,
				}
			},
		},
		{
			name: "Struct",
			args: func() args {
				type Data struct {
					Field1 string `json:"field_1"`
					Field2 int    `json:"field_2"`
				}

				data := Data{
					Field1: "field1",
					Field2: 2,
				}

				return args{
					body: toJSON(t, &data),
					ptr:  &Data{},
					want: &data,
				}
			},
		},
		{
			name: "ErrorNilBody",
			args: func() args {
				return args{
					body: nil,
				}
			},
			wantErr: true,
		},
		{
			name: "ErrorInvalidBody",
			args: func() args {
				return args{
					body: strings.NewReader("{]"),
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := NewJSON()
			args := tt.args()

			if err := j.Decode(args.body, args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			require.Equal(t, args.want, args.ptr)
		})
	}
}
