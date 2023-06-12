package bootstrap

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"fmt"
)

// Helper functions for the TemplateFuncMap var.

// Integrity
func Integrity(name string, fs embed.FS) (string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return "", err
	}
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprintf("sha384-%s", b64), nil
}
