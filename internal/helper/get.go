package helper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// Timeout is the HTTP client timeout.
	Timeout = 5 * time.Second
	// User-Agent to send with the HTTP request.
	UserAgent = "Defacto2 2024 app under construction (thanks!)"
)

// Redirect returns a new URL if the rawURL is a scene.org file URL.
func Redirect(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.Host == "scene.org" && u.Path == "/file.php" {
		// match broken legacy URLs: http://scene.org/file.php?id=299790
		return rawURL
	}
	if u.Host == "files.scene.org" {
		p := u.Path
		x := strings.Split(p, "/")
		if len(x) > 0 && x[1] == "view" {
			x[1] = "get"
			newURL := &url.URL{
				Scheme: "https",
				Host:   "files.scene.org",
				Path:   strings.Join(x, "/"),
			}
			return newURL.String()
		}
	}
	return rawURL
}

// DownloadStat returns the content length of a remote URL.
// It returns an error if the URL is invalid or the request fails.
// The content length is -1 if it is unknown.
func DownloadStat(url string) (int64, error) {
	const unknown = -1
	url = Redirect(url)
	client := http.Client{
		Timeout: Timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return unknown, err
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return unknown, err
	}
	defer res.Body.Close()
	return res.ContentLength, nil
}

// DownloadResponse contains the details of a downloaded file.
type DownloadResponse struct {
	ContentLength string // ContentLength is the size of the file in bytes.
	ContentType   string // ContentType is the MIME type of the file.
	LastModified  string // LastModified is the last modified date of the file.
	Path          string // Path is the path to the downloaded file.
}

// DownloadFile downloads a file from a remote URL and saves it to the default temp directory.
// It returns the path to the downloaded file.
func DownloadFile(url string) (DownloadResponse, error) {
	url = Redirect(url)

	// Get the remote file
	client := http.Client{
		Timeout: Timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return DownloadResponse{}, err
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return DownloadResponse{}, err
	}
	defer res.Body.Close()

	dlr := DownloadResponse{
		ContentLength: res.Header.Get("Content-Length"),
		ContentType:   res.Header.Get("Content-Type"),
		LastModified:  res.Header.Get("Last-Modified"),
	}
	// Create the file in the default temp directory
	tmpFile, err := os.CreateTemp("", "downloadfile-*")
	if err != nil {
		return DownloadResponse{}, err
	}
	defer tmpFile.Close()

	// Write the body to file
	if _, err := io.Copy(tmpFile, res.Body); err != nil {
		defer os.Remove(tmpFile.Name())
		return DownloadResponse{}, err
	}
	dlr.Path = tmpFile.Name()
	return dlr, nil
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

// RenameFileOW renames a file from oldpath to newpath.
// It returns an error if the oldpath does not exist or is a directory
// or the rename fails.
func RenameFileOW(oldpath, newpath string) error {
	st, err := os.Stat(oldpath)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("oldpath %w: %s", ErrFilePath, oldpath)
	}
	if st, err = os.Stat(newpath); err == nil {
		if st.IsDir() {
			return fmt.Errorf("newpath %w: %s", ErrFilePath, newpath)
		}
		if err = os.Remove(newpath); err != nil {
			return fmt.Errorf("newpath %w: %s", err, newpath)
		}
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
