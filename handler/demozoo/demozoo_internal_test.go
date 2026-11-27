//nolint:testpackage
package demozoo

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
)

var (
	errUnexpectedMethod = errors.New("unexpected method")
	errUnexpectedURL    = errors.New("unexpected url")
	errUnexpectedUA     = errors.New("unexpected user agent")
	errUnexpectedAuth   = errors.New("unexpected auth header")
	errNetworkDown      = errors.New("network down")
	errReadFailed       = errors.New("read failed")
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type failBody struct {
	err error
}

func (f failBody) Read([]byte) (int, error) { return 0, f.err }
func (f failBody) Close() error             { return nil }

func mockClient(wantID int, wantAuth string, status int, body string) http.Client {
	wantURL := ProdURL + strconv.Itoa(wantID)
	return http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				return nil, errUnexpectedMethod
			}
			if req.URL.String() != wantURL {
				return nil, errUnexpectedURL
			}
			if req.Header.Get("User-Agent") != helper.UserAgent {
				return nil, errUnexpectedUA
			}
			if req.Header.Get("Authorization") != "Api-Key "+wantAuth {
				return nil, errUnexpectedAuth
			}
			return &http.Response{
				StatusCode: status,
				Status:     http.StatusText(status),
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

func TestProductionGetOffline(t *testing.T) {
	t.Parallel()

	const sampleJSON = `{"id":42,"title":"Second Reality","release_date":"1993-08-01"}`

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var p Production
		code, err := p.get(t.Context(), mockClient(42, "secret", http.StatusOK, sampleJSON), "secret", 42)
		be.Err(t, err, nil)
		be.Equal(t, code, 0)
		be.Equal(t, p.ID, 42)
		be.Equal(t, p.Title, "Second Reality")
	})

	t.Run("non-200 status", func(t *testing.T) {
		t.Parallel()

		var p Production
		code, err := p.get(t.Context(), mockClient(42, "secret", http.StatusNotFound, "missing"), "secret", 42)
		be.Err(t, err, ErrStatus)
		be.Equal(t, code, http.StatusNotFound)
	})

	t.Run("id mismatch", func(t *testing.T) {
		t.Parallel()

		var p Production
		_, err := p.get(t.Context(), mockClient(42, "secret", http.StatusOK, `{"id":99}`), "secret", 42)
		be.Err(t, err, ErrSuccess)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		var p Production
		_, err := p.get(t.Context(), mockClient(42, "secret", http.StatusOK, `{bad`), "secret", 42)
		be.Err(t, err)
	})

	t.Run("null json body", func(t *testing.T) {
		t.Parallel()

		var p Production
		_, err := p.get(t.Context(), mockClient(42, "secret", http.StatusOK, `null`), "secret", 42)
		be.Err(t, err, ErrSuccess)
	})

	t.Run("transport error", func(t *testing.T) {
		t.Parallel()

		c := http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errNetworkDown
			}),
		}
		var p Production
		_, err := p.get(t.Context(), c, "secret", 42)
		be.Err(t, err, errNetworkDown)
	})

	t.Run("body read error", func(t *testing.T) {
		t.Parallel()

		c := http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       failBody{err: errReadFailed},
					Header:     make(http.Header),
				}, nil
			}),
		}
		var p Production
		_, err := p.get(t.Context(), c, "secret", 42)
		be.Err(t, err, errReadFailed)
	})
}
