package roamer

import (
	"testing"

	roamerError "github.com/SLIpros/roamer/err"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestIsDecodeError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		args   args
		want   *roamerError.DecodeError
		wantOK bool
	}{
		{
			name: "is decode error",
			args: args{
				err: &roamerError.DecodeError{},
			},
			want:   &roamerError.DecodeError{},
			wantOK: true,
		},
		{
			name: "is not decode error",
			args: args{
				err: errors.New("big bad"),
			},
			want:   nil,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := IsDecodeError(tt.args.err)
			if ok != tt.wantOK {
				t.Errorf("IsDecodeError() got1 = %v, want %v", ok, tt.wantOK)
			}

			require.Equalf(t, tt.want, got, "IsDecodeError() got = %v, want %v", got, tt.want)
		})
	}
}
