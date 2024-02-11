package roamer

import (
	"testing"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
	"github.com/stretchr/testify/require"
)

func TestIsDecodeError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		args   args
		want   rerr.DecodeError
		wantOK bool
	}{
		{
			name: "is decode error",
			args: args{
				err: rerr.DecodeError{},
			},
			want:   rerr.DecodeError{},
			wantOK: true,
		},
		{
			name: "is not decode error",
			args: args{
				err: errors.New("big bad"),
			},
			want:   rerr.DecodeError{},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := IsDecodeError(tt.args.err)
			if ok != tt.wantOK {
				t.Errorf("IsDecodeError() got1 = %v, want %v", ok, tt.wantOK)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsSliceIterationError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		args   args
		want   rerr.SliceIterationError
		wantOK bool
	}{
		{
			name: "is slice iteration error",
			args: args{
				err: rerr.SliceIterationError{},
			},
			want:   rerr.SliceIterationError{},
			wantOK: true,
		},
		{
			name: "is not slice iteration error",
			args: args{
				err: errors.New("big bad"),
			},
			want:   rerr.SliceIterationError{},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := IsSliceIterationError(tt.args.err)
			if ok != tt.wantOK {
				t.Errorf("IsSliceIterationError() got1 = %v, want %v", ok, tt.wantOK)
			}

			require.Equal(t, tt.want, got)
		})
	}
}
