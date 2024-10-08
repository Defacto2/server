package remote

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Defacto2/helper"
)

const (
	// Timeout is the HTTP client default timeout.
	Timeout = 5 * time.Second
	// User-Agent to send with the HTTP request.
	UserAgent = "Defacto2 Uploader form submission, thanks!"
)

// FixSceneOrg returns a working URL if the provided rawURL is a known,
// broken link to a scene.org file. Otherwise it returns the original URL.
//
// For example, the following rawURL:
//
//	`http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip`
//
// will return:
//
//	`https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip`
func FixSceneOrg(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.Host == "scene.org" && u.Path == "/file.php" {
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

// DownloadResponse contains the details of a downloaded file.
type DownloadResponse struct {
	ContentLength string // ContentLength is the size of the file in bytes.
	ContentType   string // ContentType is the MIME type of the file.
	LastModified  string // LastModified is the last modified date of the file.
	Path          string // Path is the path to the downloaded file.
}

// GetFile downloads a file from a remote URL and saves it to the default temp directory.
// If timeout is 0, it uses the default timeout of 5 seconds, otherwise it uses the provided timeout.
// It returns the path to the downloaded file and it should be removed after use.
func GetFile(rawURL string, timeout time.Duration) (DownloadResponse, error) {
	url := FixSceneOrg(rawURL)
	// Get the remote file
	if timeout == 0 {
		timeout = Timeout
	}
	client := http.Client{
		Timeout: timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("get file new request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("get file client do: %w", err)
	}
	defer res.Body.Close()

	download := DownloadResponse{
		ContentLength: res.Header.Get("Content-Length"),
		ContentType:   res.Header.Get("Content-Type"),
		LastModified:  res.Header.Get("Last-Modified"),
	}
	// Create the file in the default temp directory
	dst, err := os.CreateTemp(helper.TmpDir(), "get-remotefile-*")
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("get file create temp: %w", err)
	}
	defer dst.Close()

	// Write the body to file
	if _, err := io.Copy(dst, res.Body); err != nil {
		defer os.Remove(dst.Name())
		return DownloadResponse{}, fmt.Errorf("get file io copy: %w", err)
	}
	download.Path = dst.Name()
	return download, nil
}
