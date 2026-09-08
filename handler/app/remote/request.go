//nolint:nonamedreturns
package remote

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
)

const (
	// UserAgent to send with the HTTP request.
	UserAgent = "Defacto2 Uploader form submission, thanks!"

	TimeoutShort = 5 * time.Second
	TimeoutLong  = 10 * time.Second
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

	var g got
	return g.getDo(ctx, sl, rawURL)
}

type got struct {
	msg      string
	filepath string
	client   http.Client
	request  *http.Request
	response *http.Response
	dest     *os.File
}

func (g *got) clientDo() (err error) {
	// handle the response including anything unexpected
	g.response, err = g.client.Do(g.request)
	if err != nil {
		return err
	}
	if g.response == nil {
		return http.ErrBodyNotAllowed
	}
	if g.response.Body == nil {
		return ErrBodyNil
	}
	return nil
}

func (g *got) newRequest(ctx context.Context, url string) (err error) {
	g.request, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	g.request.Header.Set("User-Agent", UserAgent)

	return nil
}

func (g *got) responseDo(sl *slog.Logger, url string) (n int64, err error) {
	if g.response.StatusCode >= http.StatusBadRequest {
		sl.Info(g.msg+" http returned a unexpected status",
			slog.Int("status code", g.response.StatusCode), slog.String("url", url))
		g.cleanup(sl)
		return 0, fmt.Errorf("unexpected status %s: %w", g.response.Status, ErrStatus)
	}

	// create the file in the default temp directory
	g.dest, err = dir.CreateTemp("get-remotefile-*")
	if err != nil {
		g.cleanup(sl)
		return 0, fmt.Errorf("create temporary file: %w", err)
	}
	g.filepath = g.dest.Name()

	defer func() {
		if err := g.dest.Close(); err != nil {
			sl.Info(g.msg+" closing temporary file caused an error",
				slog.String("filename", g.filepath),
				slog.Any("error", err))
		}
	}()

	// write the http body to file
	const size = 4 * 1024
	buf := make([]byte, size)
	n, err = io.CopyBuffer(g.dest, g.response.Body, buf)
	if err != nil {
		g.cleanup(sl)
		defer func() {
			if err := os.Remove(g.filepath); err != nil {
				sl.Info(g.msg+" removing temporary file caused an error", slog.String("filename", g.filepath),
					slog.Any("error", err))
			}
		}()
		return 0, fmt.Errorf("io copy: %w", err)
	}

	return n, nil
}

// GetFile downloads a file from a remote URL and saves it to the default temp directory.
// It returns the path to the downloaded file and it should be removed after use.
//
// Because of the random silent failures of fetching remote files, the slog logger by default is verbose.
func (g *got) getDo(ctx context.Context, sl *slog.Logger, rawURL string) (Response, error) {
	g.msg = "app remote get file"

	url := ReplaceURL(sl, rawURL)

	failure := func(s, url string, err error) (Response, error) {
		sl.Info(g.msg+" "+s, slog.String("url", url), slog.Any("error", err))
		const format = "app remote request %s get file url %q: %w"
		return Response{}, fmt.Errorf(format, s, url, err)
	}

	if err := g.newRequest(ctx, url); err != nil {
		return failure("new request", url, err)
	}
	if err := g.clientDo(); err != nil {
		return failure("client do", url, err)
	}

	defer func() {
		if err := g.response.Body.Close(); err != nil {
			sl.Info(g.msg+" response body close", slog.Any("error", err))
		}
	}()

	n, err := g.responseDo(sl, url)
	if err != nil {
		return failure("response do", url, err)
	}

	sl.Info(g.msg+" copied http body to the temporary file",
		slog.String("file", g.filepath), slog.Int64("bytes", n))
	return Response{
		ContentLength: g.response.Header.Get("Content-Length"),
		ContentType:   g.response.Header.Get("Content-Type"),
		LastModified:  g.response.Header.Get("Last-Modified"),
		Path:          g.filepath,
	}, nil
}

// cleanup attempts attempt to discard and close the response body.
func (g *got) cleanup(sl *slog.Logger) {
	if g.response == nil {
		sl.Error(g.msg + " http response cannot be nil")
		return
	}

	if n, err := io.Copy(io.Discard, g.response.Body); err != nil {
		sl.Info(g.msg+" error discarding the response body", slog.Any("error", err))
	} else {
		sl.Info(g.msg+" discard response body", slog.Int64("bytes", n))
	}

	if err := g.response.Body.Close(); err != nil {
		sl.Info(g.msg+" response body close caused an error", slog.Any("error", err))
	}
}
