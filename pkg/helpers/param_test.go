package helpers_test

import (
	"strconv"
	"testing"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/stretchr/testify/assert"
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

func TestObfuscates(t *testing.T) {
	keys := []int{1, 1000, 1236346, -123, 0}
	for _, key := range keys {
		s := helpers.ObfuscateParam(strconv.Itoa(key))
		assert.Equal(t, key, helpers.Deobfuscate(s))
	}
}
