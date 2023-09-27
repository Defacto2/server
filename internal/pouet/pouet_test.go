package pouet_test

import (
	"testing"

	"github.com/Defacto2/server/internal/pouet"
	"github.com/stretchr/testify/assert"
)

func TestStars(t *testing.T) {
	type args struct {
		up   uint64
		meh  uint64
		down uint64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"0", args{0, 0, 0}, 0},
		{"1 up", args{1, 0, 0}, 5},
		{"1 meh", args{0, 1, 0}, 3},
		{"1 down", args{0, 0, 1}, 1},
		{"2 below avg", args{0, 1, 1}, 2},
		{"1s", args{1, 1, 1}, 3},
		{"1,1,0", args{1, 1, 0}, 4},
		{"2,1,0", args{2, 1, 0}, 4.5},
		{"3,1,0", args{3, 1, 0}, 4.5},
		{"7,1,0", args{7, 1, 0}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want,
				pouet.Stars(tt.args.up, tt.args.meh, tt.args.down))
		})
	}
}
