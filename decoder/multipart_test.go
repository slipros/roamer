package decoder

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type multipartFormDataTestData struct {
	Image      string  `multipart:"image_url"`
	ImageAsPtr *string `multipart:"image_url"`
	ImageAsURL url.URL `multipart:"image_url"`
	NoKey      string  `multipart:"no_key"`
	NoKeyAsPtr *string `multipart:"no_key"`

	TextFile      MultipartFile  `multipart:"text_file"`
	OtherTextFile *MultipartFile `multipart:"other_text_file"`
	NoFile        MultipartFile  `multipart:"no_file"`
	NoFileAsPtr   *MultipartFile `multipart:"no_file"`
	AllFiles      MultipartFiles `multipart:",allfiles"`
}

func TestNewMultipartFormData(t *testing.T) {
	m := NewMultipartFormData()
	require.NotNil(t, m)
	require.Equal(t, ContentTypeMultipartFormData, m.ContentType())

	m = NewMultipartFormData(WithMaxMemory(1000))
	require.NotNil(t, m)
	require.Equal(t, ContentTypeMultipartFormData, m.ContentType())
	require.Equal(t, int64(1000), m.maxMemory)

	m = NewMultipartFormData(WithContentType[*MultipartFormData]("test"))
	require.NotNil(t, m)
	require.Equal(t, "test", m.ContentType())

	m = NewMultipartFormData(WithSkipFilled[*MultipartFormData](false))
	require.NotNil(t, m)
	require.Equal(t, false, m.skipFilled)
}

func TestMultipartFormData_Decode(t *testing.T) {
	type args struct {
		r    *http.Request
		ptr  *multipartFormDataTestData
		want *multipartFormDataTestData
	}
	tests := []struct {
		name    string
		args    func() args
		wantErr bool
	}{
		{
			name: "success",
			args: func() args {
				r, ptr, want := prepareMultipartFormDataArgs()

				return args{
					r:    r,
					ptr:  ptr,
					want: want,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			m := NewMultipartFormData()
			if err := m.Decode(args.r, args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, "http://some.url", args.ptr.Image)
			require.NotNil(t, args.ptr.ImageAsPtr)
			require.Equal(t, "http://some.url", *args.ptr.ImageAsPtr)
			require.Equal(t, "http://some.url", args.ptr.ImageAsURL.String())

			require.Empty(t, args.ptr.NoKey)
			require.Nil(t, args.ptr.NoKeyAsPtr)

			require.Equal(t, "text_file", args.ptr.TextFile.Key)
			require.NotNil(t, args.ptr.TextFile.File)
			require.NotNil(t, args.ptr.TextFile.Header)

			require.NotNil(t, args.ptr.OtherTextFile)
			require.Equal(t, "other_text_file", args.ptr.OtherTextFile.Key)
			require.NotNil(t, args.ptr.OtherTextFile.File)
			require.NotNil(t, args.ptr.OtherTextFile.Header)

			require.Empty(t, args.ptr.NoFile)
			require.Nil(t, args.ptr.NoFileAsPtr)

			require.NotEmpty(t, args.ptr.AllFiles)
			require.Equal(t, 2, len(args.ptr.AllFiles))
		})
	}
}

func BenchmarkMultipartFormData_Decode(b *testing.B) {
	r, ptr, _ := prepareMultipartFormDataArgs()

	m := NewMultipartFormData(WithSkipFilled[*MultipartFormData](false))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := m.Decode(r, ptr); err != nil {
			b.Fatal(err)
		}
	}
}

func prepareMultipartFormDataArgs() (r *http.Request, ptr *multipartFormDataTestData, want *multipartFormDataTestData) {
	imageValue := "http://some.url"

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	err := w.WriteField("image_url", imageValue)
	if err != nil {
		panic(err)
	}

	textFileWriter, err := w.CreateFormFile("text_file", "hello.txt")
	if err != nil {
		panic(err)
	}

	_, err = textFileWriter.Write([]byte("Hello"))
	if err != nil {
		panic(err)
	}

	otherTextFileWriter, err := w.CreateFormFile("other_text_file", "other_hello.txt")
	if err != nil {
		panic(err)
	}

	_, err = otherTextFileWriter.Write([]byte("Hello2"))
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	contentType := w.FormDataContentType()

	// prepare want
	dummyReq, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(b.Bytes()))
	if err != nil {
		panic(err)
	}

	dummyReq.Header.Add("Content-Type", contentType)

	err = dummyReq.ParseMultipartForm(defaultMultipartFormDataMaxMemory)
	if err != nil {
		panic(err)
	}

	file, header, err := dummyReq.FormFile("text_file")
	if err != nil {
		panic(err)
	}

	textFile := MultipartFile{
		Key:    "text_file",
		File:   file,
		Header: header,
	}

	file, header, err = dummyReq.FormFile("other_text_file")
	if err != nil {
		panic(err)
	}

	otherTextFile := MultipartFile{
		Key:    "other_text_file",
		File:   file,
		Header: header,
	}

	imageAsURL, err := url.Parse(imageValue)
	if err != nil {
		panic(err)
	}

	want = &multipartFormDataTestData{
		Image:         imageValue,
		ImageAsPtr:    &imageValue,
		ImageAsURL:    *imageAsURL,
		NoKey:         "",
		NoKeyAsPtr:    nil,
		TextFile:      textFile,
		OtherTextFile: &otherTextFile,
		NoFile:        MultipartFile{},
		NoFileAsPtr:   nil,
		AllFiles: MultipartFiles{
			textFile,
			otherTextFile,
		},
	}
	//

	r, err = http.NewRequest(http.MethodPost, requestURL, &b)
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", contentType)

	return r, &multipartFormDataTestData{}, want
}
