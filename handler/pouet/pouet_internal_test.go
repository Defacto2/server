package pouet

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

func mockClient(status int, body string) http.Client {
	const wantID = 42
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
			return &http.Response{
				StatusCode: status,
				Status:     http.StatusText(status),
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

const sampleProdJSON = `{
	"success": true,
	"prod": {
		"id": "42",
		"name": "Second Reality",
		"releaseDate": "1993-08-01",
		"voteup": "100",
		"votepig": "10",
		"votedown": "5",
		"voteavg": "0.85",
		"download": "https://example.com/2ndreal.zip",
		"demozoo": "123",
		"groups": [{"id": "1", "name": "Future Crew"}],
		"platforms": {"67": {"name": "MS-Dos", "slug": "msdos"}},
		"types": ["demo"],
		"downloadLinks": [{"type": "mirror", "link": "https://mirror.example.com/2ndreal.zip"}]
	}
}`

func TestResponseGetOffline(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var r Response
		code, err := r.get(t.Context(), mockClient(http.StatusOK, sampleProdJSON), 42)
		be.Err(t, err, nil)
		be.Equal(t, code, 0)
		be.True(t, r.Success)
		be.Equal(t, r.Prod.Title, "Second Reality")
	})

	t.Run("non-200 status", func(t *testing.T) {
		t.Parallel()

		var r Response
		code, err := r.get(t.Context(), mockClient(http.StatusNotFound, "not found"), 42)
		be.Err(t, err, ErrStatusCode)
		be.Equal(t, code, http.StatusNotFound)
	})

	t.Run("success false", func(t *testing.T) {
		t.Parallel()

		var r Response
		_, err := r.get(t.Context(), mockClient(http.StatusOK, `{"success":false}`), 42)
		be.Err(t, err, ErrSuccess)
	})

	t.Run("null json", func(t *testing.T) {
		t.Parallel()

		var r Response
		_, err := r.get(t.Context(), mockClient(http.StatusOK, `null`), 42)
		be.Err(t, err, ErrSuccess)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		var r Response
		_, err := r.get(t.Context(), mockClient(http.StatusOK, `{bad`), 42)
		be.Err(t, err)
	})

	t.Run("transport error", func(t *testing.T) {
		t.Parallel()

		c := http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errNetworkDown
			}),
		}
		var r Response
		_, err := r.get(t.Context(), c, 42)
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
		var r Response
		_, err := r.get(t.Context(), c, 42)
		be.Err(t, err, errReadFailed)
	})
}

func TestProductionGetOffline(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var p Production
		code, err := p.get(t.Context(), mockClient(http.StatusOK, sampleProdJSON), 42)
		be.Err(t, err, nil)
		be.Equal(t, code, 0)
		be.Equal(t, p.ID, 42)
		be.Equal(t, p.Title, "Second Reality")
		be.Equal(t, p.Demozoo, "123")
		be.True(t, p.Valid)
		be.Equal(t, len(p.Groups), 1)
		be.Equal(t, len(p.Links), 1)
	})

	t.Run("http error", func(t *testing.T) {
		t.Parallel()

		var p Production
		code, err := p.get(t.Context(), mockClient(http.StatusServiceUnavailable, ""), 42)
		be.Err(t, err, ErrStatusCode)
		be.Equal(t, code, http.StatusServiceUnavailable)
	})

	t.Run("invalid prod id", func(t *testing.T) {
		t.Parallel()

		var p Production
		_, err := p.get(t.Context(), mockClient(http.StatusOK, `{"success":true,"prod":{"id":"abc"}}`), 42)
		be.Err(t, err, strconv.ErrSyntax)
	})
}

func TestVotesOffline(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var v Votes
		err := v.votes(t.Context(), mockClient(http.StatusOK, sampleProdJSON), 42)
		be.Err(t, err, nil)
		be.Equal(t, v.ID, 42)
		be.Equal(t, v.VotesUp, uint64(100))
		be.Equal(t, v.VotesMeh, uint64(10))
		be.Equal(t, v.VotesDown, uint64(5))
		be.Equal(t, v.VotesAvg, 0.85)
		be.Equal(t, v.Stars, 4.5)
	})

	tests := []struct {
		name    string
		body    string
		wantErr error
	}{
		{"success false", `{"success":false}`, ErrSuccess},
		{"bad id", `{"success":true,"prod":{"id":"x"}}`, strconv.ErrSyntax},
		{"bad up", `{"success":true,"prod":{"id":"1","voteup":"x"}}`, strconv.ErrSyntax},
		{"bad meh", `{"success":true,"prod":{"id":"1","voteup":"1","votepig":"x"}}`, strconv.ErrSyntax},
		{"bad down", `{"success":true,"prod":{"id":"1","voteup":"1","votepig":"1","votedown":"x"}}`, strconv.ErrSyntax},
		{"bad avg", `{"success":true,"prod":{"id":"1","voteup":"1","votepig":"1","votedown":"1","voteavg":"x"}}`, strconv.ErrSyntax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var v Votes
			err := v.votes(t.Context(), mockClient(http.StatusOK, tt.body), 42)
			be.Err(t, err, tt.wantErr)
		})
	}
}
