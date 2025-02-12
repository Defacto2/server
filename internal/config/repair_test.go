package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemove(t *testing.T) {
	tmpdiz := filepath.Join(t.TempDir(), "file_id.diz")
	tmptxt := filepath.Join(t.TempDir(), "readme.txt")
	tests := []struct {
		name  string
		diz   string
		txt   string
		want  string
		setup func() error
	}{
		{
			name: "Is a diz file",
			diz:  tmpdiz,
			txt:  tmptxt,
			want: "readme.txt",
			setup: func() error {
				data := []byte("FILE_ID.DIZ content")
				if err := os.WriteFile(tmpdiz, data, 0644); err != nil {
					return err
				}
				return os.WriteFile(tmptxt, data, 0644)
			},
		},
		{
			name: "Is too wide file",
			diz:  tmpdiz,
			txt:  tmptxt,
			want: "file_id.diz",
			setup: func() error {
				const x = "1234567890"
				data := []byte(strings.Repeat(x, 100))
				if err := os.WriteFile(tmpdiz, data, 0644); err != nil {
					return err
				}
				return os.WriteFile(tmptxt, data, 0644)
			},
		},
		{
			name: "Is too long file",
			diz:  tmpdiz,
			txt:  tmptxt,
			want: "file_id.diz",
			setup: func() error {
				const x = "1234567890\n"
				data := []byte{}
				for range 15 {
					data = append(data, x...)
				}
				if err := os.WriteFile(tmpdiz, data, 0644); err != nil {
					return err
				}
				return os.WriteFile(tmptxt, data, 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}
			got, err := Remove(tt.diz, tt.txt)
			if err != nil {
				t.Errorf("Remove() error = %v, wantErr %v", err, false)
				return
			}
			if got != tt.want {
				t.Errorf("Remove() got = %v, want %v", got, tt.want)
			}
		})
	}
}
