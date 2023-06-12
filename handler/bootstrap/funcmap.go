package bootstrap

import (
	"crypto/sha512"
	"embed"
	"fmt"
)

// Helper functions for the TemplateFuncMap var.

// Integrity
func Integrity(name string, fs embed.FS) string {
	b, err := fs.ReadFile(name)
	if err != nil {
		return ""
	}
	sum := sha512.Sum384(b)
	s := string(sum[:])
	if s == "" {
		return ""
	}
	return fmt.Sprintf("sha384-%s", s)
}
