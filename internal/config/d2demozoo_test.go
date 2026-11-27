package config_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestDemozooAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		val       config.DemozooAuth
		wantLog   string
		wantHelp  string
		wantIssue string
	}{
		{"empty", "", "", "demozoo.org", ""},
		{"short", "12345678", "xxxxx", "", "too short"},
		{"no dot", "1234567890", "xxxxx", "", "period"},
		{"valid", "12345678.abcdefgh", "xxxxx", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, tt.val.LogValue().String(), tt.wantLog)
			if tt.wantHelp == "" {
				be.Equal(t, tt.val.Help(), "")
			} else {
				be.True(t, strings.Contains(tt.val.Help(), tt.wantHelp))
			}
			if tt.wantIssue == "" {
				be.Equal(t, tt.val.Issue(), "")
			} else {
				be.True(t, strings.Contains(tt.val.Issue(), tt.wantIssue))
			}
		})
	}
}
