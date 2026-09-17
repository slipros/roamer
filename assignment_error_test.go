package roamer

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/slipros/assign"
	"github.com/slipros/roamer/decoder"
	rerr "github.com/slipros/roamer/err"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/mockroamer"
	"github.com/slipros/roamer/parser"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoamer_Parse_AssignmentError(t *testing.T) {
	t.Parallel()

	// region data provider
	tests := []struct {
		name   string
		tag    string
		value  string
		newPtr func() any
		cause  error
	}{
		{
			name: "query scalar", tag: "query", value: "bad", cause: strconv.ErrSyntax,
			newPtr: func() any {
				return new(struct {
					Value uint32 `query:"value"`
				})
			},
		},
		{
			name: "query pointer", tag: "query", value: "bad", cause: strconv.ErrSyntax,
			newPtr: func() any {
				return new(struct {
					Value *uint32 `query:"value"`
				})
			},
		},
		{
			name: "query overflow", tag: "query", value: "18446744073709551616", cause: strconv.ErrRange,
			newPtr: func() any {
				return new(struct {
					Value *uint32 `query:"value"`
				})
			},
		},
		{
			name: "path", tag: "path", value: "bad", cause: strconv.ErrSyntax,
			newPtr: func() any {
				return new(struct {
					Value uint32 `path:"value"`
				})
			},
		},
		{
			name: "header", tag: "header", value: "bad", cause: strconv.ErrSyntax,
			newPtr: func() any {
				return new(struct {
					Value bool `header:"Value"`
				})
			},
		},
		{
			name: "internal parser", tag: "profile", value: "bad", cause: strconv.ErrSyntax,
			newPtr: func() any {
				return new(struct {
					Value uint32 `profile:"value"`
				})
			},
		},
		{
			name: "unsupported destination", tag: "query", value: "bad", cause: assign.ErrNotSupported,
			newPtr: func() any {
				return new(struct {
					Value chan int `query:"value"`
				})
			},
		},
	}
	// endregion

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/?value="+tt.value, http.NoBody)
			req.Header.Set("Value", tt.value)
			profile := mockroamer.NewParser(t)
			profile.EXPECT().Tag().Return("profile")
			if tt.tag == "profile" {
				profile.EXPECT().Parse(req, mock.Anything, mock.Anything).Return(tt.value, true).Once()
			}
			r := NewRoamer(WithParsers(
				parser.NewQuery(),
				parser.NewPath(func(_ *http.Request, _ string) (string, bool) { return tt.value, true }),
				parser.NewHeader(),
				profile,
			))

			err := r.Parse(req, tt.newPtr())
			require.Error(t, err)
			assignmentErr, ok := IsAssignmentError(errors.WithMessage(err, "request failed"))
			require.True(t, ok, "assignment must expose its parser source: %v", err)
			require.Equal(t, "Value", assignmentErr.Field)
			require.Equal(t, tt.tag, assignmentErr.Tag)
			require.ErrorIs(t, err, tt.cause)
			require.Equal(t, assignmentErr.Err.Error(), assignmentErr.Error())
			var typed rerr.AssignmentError
			require.ErrorAs(t, err, &typed)
			require.Equal(t, assignmentErr, typed)
			if !errors.Is(err, assign.ErrNotSupported) {
				var numberErr *strconv.NumError
				require.ErrorAs(t, err, &numberErr)
				require.Equal(t, tt.value, numberErr.Num)
			}
		})
	}
}

func TestRoamer_Parse_NonAssignmentError(t *testing.T) {
	t.Parallel()

	// region data provider
	tests := []struct {
		name   string
		newPtr func() any
		body   string
	}{
		{
			name: "default", body: `{}`,
			newPtr: func() any {
				return new(struct {
					Value uint32 `default:"bad"`
				})
			},
		},
		{
			name: "formatter", body: `{}`,
			newPtr: func() any {
				return new(struct {
					Value string `query:"value" string:"unknown"`
				})
			},
		},
		{
			name: "body decoding", body: `{"value":`,
			newPtr: func() any {
				return new(struct {
					Value uint32 `json:"value"`
				})
			},
		},
		{
			name: "after parse", body: `{}`,
			newPtr: func() any { return new(errorAfterParser) },
		},
	}
	// endregion

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/?value=valid", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r := NewRoamer(WithParsers(parser.NewQuery()), WithDecoders(decoder.NewJSON()),
				WithFormatters(formatter.NewString()))
			err := r.Parse(req, tt.newPtr())
			require.Error(t, err)
			_, ok := IsAssignmentError(err)
			require.False(t, ok, "unrelated failures must not acquire parser provenance")
		})
	}
}

func TestRoamer_Parse_PointerAssignment_Successfully(t *testing.T) {
	t.Parallel()

	// region data provider
	tests := []struct {
		name   string
		target string
		value  uint32
		set    bool
	}{
		{name: "absent", target: "/"},
		{name: "present", target: "/?pageNumber=2", value: 2, set: true},
	}
	// endregion

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var data struct {
				PageNumber *uint32 `query:"pageNumber"`
			}
			r := NewRoamer(WithParsers(parser.NewQuery()))
			err := r.Parse(httptest.NewRequest(http.MethodGet, tt.target, http.NoBody), &data)
			require.NoError(t, err)
			_, ok := IsAssignmentError(err)
			require.False(t, ok)
			if !tt.set {
				require.Nil(t, data.PageNumber)

				return
			}
			require.NotNil(t, data.PageNumber)
			require.Equal(t, tt.value, *data.PageNumber)
		})
	}
}
