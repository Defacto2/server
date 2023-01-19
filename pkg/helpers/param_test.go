package helpers_test

import (
	"reflect"
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

func TestPages(t *testing.T) {
	type args struct {
		sum    int
		limit  int
		offset int
	}
	//onePage := []uint{1}
	const k, h = 1000, 100
	tests := []struct {
		name string
		args args
		want []uint
	}{
		// {"-1", args{-1, -1, -1}, onePage},
		// {"too few records", args{200, 500, 1}, onePage},
		// {"page 1", args{k, k, 1}, onePage},
		// {"page 1.1", args{11, 10, 1}, []uint{1, 2}},
		// {"page 3", args{21, 10, 1}, []uint{1, 2, 3}},
		// {"page 4", args{31, 10, 1}, []uint{1, 2, 3, 4}},
		// {"page 5", args{41, 10, 1}, []uint{1, 2, 3, 4, 5}},
		// {"hundres", args{410, 10, 1}, []uint{1, 2, 3, 4, 5}},
		// {"page 2", args{k, h, 101}, []uint{1, 2, 3, 4, 10}},
		// {"page 3", args{k, h, 243}, []uint{1, 2, 3, 4, 5, 10}},
		// {"page 4", args{k, h, 319}, []uint{1, 3, 4, 5, 6, 10}},
		// {"page 5", args{k, h, 423}, []uint{1, 4, 5, 6, 7, 10}},
		// {"page 6", args{k, h, 501}, []uint{1, 5, 6, 7, 8, 10}},
		{"page 7", args{k, h, 665}, []uint{1, 6, 7, 8, 9, 10}},
		{"page 8", args{k, h, 777}, []uint{1, 6, 7, 8, 9, 10}},
		{"page 9", args{k, h, 888}, []uint{1, 6, 7, 8, 9, 10}},
		{"page 10", args{k, h, k}, []uint{1, 6, 7, 8, 9, 10}},
		{"100s", args{410, 10, 1}, []uint{1, 2, 3, 4, 5, 41}},
		{"100s", args{410, 10, 100}, []uint{1, 9, 10, 11, 12, 41}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helpers.Pages(tt.args.sum, tt.args.limit, tt.args.offset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pages() = %v, want %v", got, tt.want)
			}
		})
	}
}
