package helpers_test

import (
	"testing"

	"github.com/Defacto2/server/pkg/helpers"
)

func TestPageCount(t *testing.T) {
	type args struct {
		sum   int
		limit int
	}
	tests := []struct {
		name  string
		args  args
		want1 uint
	}{
		{"-1", args{-1, -1}, 0},
		{"0", args{0, 500}, 0},
		{"1", args{1, 500}, 1},
		{"500", args{500, 750}, 1},
		{"750", args{750, 500}, 2},
		{"1k", args{1000, 500}, 2},
		{"1001", args{1001, 500}, 3},
		{"want 10", args{1000, 100}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpers.PageCount(tt.args.sum, tt.args.limit)
			if got != tt.want1 {
				t.Errorf("PageCount() got = %v, want %v", got, tt.want1)
			}
		})
	}
}
