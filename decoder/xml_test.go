package decoder

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewXML(t *testing.T) {
	d := NewXML()
	require.NotNil(t, d)
	require.Equal(t, ContentTypeXML, d.ContentType())
}

func TestXML_Decode(t *testing.T) {
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
					body: toXML(t, &data),
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
					body: strings.NewReader("<></>"),
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewXML()
			args := tt.args()

			if err := x.Decode(args.body, args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			require.Equal(t, args.want, args.ptr)
		})
	}
}
