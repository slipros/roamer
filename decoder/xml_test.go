package decoder

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestXML_Tag(t *testing.T) {
	x := NewXML()
	assert.Equal(t, TagXML, x.Tag())
}

func TestXML_Decode_Successfully(t *testing.T) {
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

	x := NewXML()
	ptr := &Data{}
	err = x.Decode(req, ptr)

	require.NoError(t, err)
	require.Equal(t, &data, ptr)
}

func TestXML_Decode_Failure(t *testing.T) {
	type args struct {
		req *http.Request
		ptr any
	}
	tests := []struct {
		name string
		args func() args
	}{
		{
			name: "Error request body is nil",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				return args{
					req: req,
				}
			},
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewXML()
			args := tt.args()

			err := x.Decode(args.req, args.ptr)
			require.Error(t, err)
		})
	}
}
