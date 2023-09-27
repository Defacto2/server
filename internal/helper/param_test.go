package helper_test

import (
	"testing"

	"github.com/Defacto2/server/pkg/helper"
	"github.com/stretchr/testify/assert"
)

func TestPageCount(t *testing.T) {
	type args struct {
		sum   int
		limit int
	}
	tests := []struct {
		name string
		args args
		want uint
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
			assert.Equal(t, tt.want, helper.PageCount(tt.args.sum, tt.args.limit))
		})
	}
}

func TestObfuscates(t *testing.T) {
	keys := []int{1, 1000, 1236346, -123, 0}
	for _, key := range keys {
		s := helper.ObfuscateID(int64(key))
		assert.Equal(t, key, helper.DeobfuscateID(s))
	}
}

// https://defacto2.net/f/ab27b2e

func TestDeobfuscateURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   int
	}{
		{"record", "https://defacto2.net/f/ab27b2e", 13526},
		{"download", "https://defacto2.net/d/ab27b2e", 13526},
		{"query", "https://defacto2.net/f/ab27b2e?blahblahblah", 13526},
		{"typo", "https://defacto2.net/f/ab27b2", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.DeobfuscateURL(tt.rawURL))
		})
	}
}

func TestSearchTerm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		// {"empty", "", []string{}},
		// {"spaces", "   ", []string{}},
		// {"one", "one", []string{"one"}},
		// {"two", "one two", []string{"one", "two"}},
		// {"three", "one two three", []string{"one", "two", "three"}},
		{"quotes", `"one two" three`, []string{"one two", "three"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.SearchTerm(tt.input))
		})
	}
}
