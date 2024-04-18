package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/volatiletech/null/v8"
)

func TestValidMagic(t *testing.T) {
	tests := []struct {
		name      string
		mediatype string
		expected  null.String
	}{
		{
			name:      "Empty mediatype",
			mediatype: "",
			expected:  null.String{String: "", Valid: false},
		},
		{
			name:      "Invalid mediatype",
			mediatype: "application/json",
			expected:  null.StringFrom("application/json"),
		},
		{
			name:      "Valid mediatype",
			mediatype: "image/jpeg",
			expected:  null.StringFrom("image/jpeg"),
		},
		{
			name:      "Valid mediatype",
			mediatype: "application/x-msdos-program",
			expected:  null.StringFrom("application/x-msdos-program"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.ValidMagic(tt.mediatype)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}
