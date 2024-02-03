package helper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// Timeout is the HTTP client timeout.
	Timeout = 5 * time.Second
	// User-Agent to send with the HTTP request.
	UserAgent = "Defacto2 2024 app under construction (thanks!)"
)

// DownloadFile downloads a file from a remote URL and saves it to the default temp directory.
// It returns the path to the downloaded file.
func DownloadFile(url string) (string, error) {
	// Get the remote file
	client := http.Client{
		Timeout: Timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Create the file in the default temp directory
	tmpFile, err := os.CreateTemp("", "downloadfile-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Write the body to file
	if _, err := io.Copy(tmpFile, res.Body); err != nil {
		defer os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}

// RenameFile renames a file from oldpath to newpath.
// It returns an error if the oldpath does not exist or is a directory,
// newpath already exists, or the rename fails.
func RenameFile(oldpath, newpath string) error {
	st, err := os.Stat(oldpath)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("oldpath %w: %s", ErrFilePath, oldpath)
	}
	if _, err = os.Stat(newpath); err == nil {
		return fmt.Errorf("newpath %w: %s", ErrExistPath, newpath)
	}
	if err := os.Rename(oldpath, newpath); err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && linkErr.Err.Error() == "invalid cross-device link" {
			return RenameCrossDevice(oldpath, newpath)
		}
		return err
	}
	return nil
}

// RenameCrossDevice is a workaround for renaming files across different devices.
// A cross device can also be a different file system such as a Docker volume.
func RenameCrossDevice(oldpath, newpath string) error {
	src, err := os.Open(oldpath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(newpath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	fi, err := os.Stat(oldpath)
	if err != nil {
		defer os.Remove(newpath)
		return err
	}
	if err = os.Chmod(newpath, fi.Mode()); err != nil {
		defer os.Remove(newpath)
		return err
	}
	defer os.Remove(oldpath)
	return nil
}
