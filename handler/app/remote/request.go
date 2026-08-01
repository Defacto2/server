package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
)

const (
	// UserAgent to send with the HTTP request.
	UserAgent = "Defacto2 Uploader form submission, thanks!"

	TimeoutShort = 5 * time.Second
	TimeoutLong  = 10 * time.Second
)

var (
	ErrBodyNil = errors.New("body is nil")
	ErrTimeout = errors.New("the time duration is out of range")
	ErrStatus  = errors.New("wrong status code")
)

// Response contains the details of a fetched downloaded file.
type Response struct {
	ContentLength string // ContentLength is the size of the file in bytes.
	ContentType   string // ContentType is the MIME type of the file.
	LastModified  string // LastModified is the last modified date of the file.
	Path          string // Path is the path to the downloaded file.
}

// GetFile downloads a file from a remote URL and saves it to the default temp directory.
// The timeout is used both for the context and the http client and should be either [TimeoutShort]
// or [TimeoutLong]. There is a timeout sanity check of 2 to 60 seconds.
//
// The returned [Response.Path] is the path to the downloaded file and it should be removed after use.
func GetFile(ctx context.Context, sl *slog.Logger, timeout time.Duration, rawURL string) (Response, error) {
	if err := nils.Check(ctx, sl); err != nil {
		return Response{}, fmt.Errorf("request get file check: %w", err)
	}
	const minimum = 2
	if timeout.Seconds() < minimum {
		timeout = minimum
	}
	const maximum = 60
	if timeout.Seconds() > maximum {
		timeout = maximum
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client := http.Client{} //nolint:exhaustruct
	return getFile(ctx, sl, rawURL, client)
}

// GetFile downloads a file from a remote URL and saves it to the default temp directory.
// It returns the path to the downloaded file and it should be removed after use.
//
// Because of the random silent failures of fetching remote files, the slog logger by default is verbose.
func getFile( //nolint:funlen
	ctx context.Context, sl *slog.Logger, rawURL string, client http.Client,
) (Response, error) {
	const msg = "app remote get file"
	const format = "app remote request %s get file url %q: %w"
	url := FixURL(sl, rawURL)

	// handle and log errors and failures
	none := Response{ContentLength: "", ContentType: "", LastModified: "", Path: ""}
	failure := func(s, url string, err error) (Response, error) {
		sl.Info(msg+" "+s,
			slog.String("url", url), slog.Any("error", err))
		return none, fmt.Errorf(format, s, url, err)
	}

	// get request with context
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return failure("new request", url, err)
	}
	request.Header.Set("User-Agent", UserAgent)

	// handle the response including anything unexpected
	response, err := client.Do(request)
	if err != nil {
		return failure("client do", url, err)
	}
	if response == nil {
		return none, http.ErrBodyNotAllowed
	}
	if response.Body == nil {
		return failure("empty body with status "+response.Status, url, err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			sl.Info(msg+" response body close caused an error", slog.Any("error", err))
		}
	}()
	if response.StatusCode >= http.StatusBadRequest {
		sl.Info(msg+" http returned a unexpected status",
			slog.Int("status code", response.StatusCode), slog.String("url", url))
		cleanup(sl, msg, response)
		return none, fmt.Errorf(format, "unexpected status "+response.Status, url, ErrStatus)
	}

	// create the file in the default temp directory
	dir := helper.TmpDir()
	dst, err := os.CreateTemp(dir, "get-remotefile-*")
	if err != nil {
		cleanup(sl, msg, response)
		return none, fmt.Errorf(format, "create temporary file in the directory "+dir, url, err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			sl.Info(msg+" closing temporary file caused an error",
				slog.String("directory", dir),
				slog.String("filename", dst.Name()),
				slog.Any("error", err))
		}
	}()

	// write the http body to file
	const size = 4 * 1024
	buf := make([]byte, size)
	n, err := io.CopyBuffer(dst, response.Body, buf)
	if err != nil {
		cleanup(sl, msg, response)
		defer func() {
			if err := os.Remove(dst.Name()); err != nil {
				sl.Info(msg+" removing temporary file caused an error",
					slog.String("directory", dir),
					slog.String("filename", dst.Name()),
					slog.Any("error", err))
			}
		}()
		return none, fmt.Errorf("%s io copy: %w", format, err)
	}
	sl.Info(msg+" copied http body to the temporary file",
		slog.String("file", dst.Name()), slog.Int64("bytes", n))
	return Response{
		ContentLength: response.Header.Get("Content-Length"),
		ContentType:   response.Header.Get("Content-Type"),
		LastModified:  response.Header.Get("Last-Modified"),
		Path:          dst.Name(),
	}, nil
}

// cleanup attempts attempt to discard and close the response body.
func cleanup(sl *slog.Logger, msg string, r *http.Response) {
	if r == nil {
		sl.Error(msg + " http response cannot be nil")
		return
	}
	if n, err := io.Copy(io.Discard, r.Body); err != nil {
		sl.Info(msg+" error discarding the response body", slog.Any("error", err))
	} else {
		sl.Info(msg+" discard response body", slog.Int64("bytes", n))
	}
	if err := r.Body.Close(); err != nil {
		sl.Info(msg+" response body close caused an error", slog.Any("error", err))
	}
}

// FixURL returns a valid URL if the provided rawURL is a known broken link to a scene.org file.
// Otherwise the rawURL is returned.
//
// For example, the following rawURL:
//
//	`http://files.scene.org/view/demos/groups/trsi/ms-dos/trsiscxt.zip`
//
// will return:
//
//	`https://files.scene.org/get/demos/groups/trsi/ms-dos/trsiscxt.zip`
func FixURL(sl *slog.Logger, rawURL string) string {
	if sl == nil {
		sl = logs.Discard()
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		sl.Info("FixURL url would not be parsed", slog.String("url", rawURL), slog.Any("error", err))
		return rawURL
	}
	if u.Host == "scene.org" && u.Path == "/file.php" {
		return rawURL
	}
	p := u.Path
	if u.Host == "files.scene.org" {
		x := strings.Split(p, "/")
		if len(x) > 0 && x[1] == "view" {
			url := sceneorg(x).String()
			sl.Info("FixURL refactored the url", slog.String("old", rawURL), slog.String("new", url))
			return url
		}
	}
	if u.Host == "ftp.scene.org" || u.Host == "ftp.pl.scene.org" || u.Host == "ftp.no.scene.org" {
		x := strings.Split(p, "/")
		if len(x) > 0 && x[1] == "pub" {
			url := sceneorg(x).String()
			sl.Info("FixURL refactored the url", slog.String("old", rawURL), slog.String("new", url))
			return url
		}
	}
	if u.Host == "sceneorg.retropc.se" || u.Host == "mirror.netcologne.de" {
		x := strings.Split(p, "/")
		if len(x) > 0 && x[1] == "scene.org" {
			x = slices.Insert(x, 1, "get")
			url := sceneorg(x).String()
			sl.Info("FixURL refactored the url", slog.String("old", rawURL), slog.String("new", url))
			return url
		}
	}
	if u.Host == "discmaster.textfiles.com" {
		x := strings.Split(p, "/")
		if len(x) > 0 && x[1] == "view" {
			x[1] = "file"
			url := refactor(u.Host, x).String()
			sl.Info("FixURL refactored the url", slog.String("old", rawURL), slog.String("new", url))
			return url
		}
	}
	return rawURL
}

func sceneorg(x []string) *url.URL {
	const minimum = 2
	if len(x) < minimum {
		return &url.URL{} //nolint:exhaustruct
	}
	x[1] = "get"
	return refactor("files.scene.org", x)
}

func refactor(host string, x []string) *url.URL {
	return &url.URL{
		Scheme:      "https",
		Opaque:      "",
		User:        nil,
		Host:        host,
		Path:        strings.Join(x, "/"),
		RawQuery:    "",
		Fragment:    "",
		RawPath:     "",
		RawFragment: "",
		ForceQuery:  false,
		OmitHost:    false,
	}
}
