package handler_test

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

// TestHandler mutates package-level app.Caching via Handler -> RouteFS, so it cannot run in parallel.
func TestHandler(t *testing.T) { //nolint:paralleltest
	db, err := sql.Open("pgx", "")
	be.Err(t, err, nil)
	t.Cleanup(func() {
		_ = db.Close()
	})
	root := testutil.OpenRoot(t)

	c := config.Config{
		Compression: true,
		ProdMode:    true,
		Quiet:       true,
	}
	h := handler.Server{
		Environment: c,
		Public:      root.FS(),
		View:        root.FS(),
	}

	e := h.Handler(t.Context(), testLog, db)
	var hasRoutes, hasErrs bool
	for _, rt := range e.Router().Routes() {
		switch rt.Path {
		case "/editor/routes":
			hasRoutes = true
		case "/errs/ae":
			hasErrs = true
		}
	}
	be.True(t, hasRoutes)
	be.True(t, hasErrs)
}

func TestStart(t *testing.T) {
	t.Parallel()

	h := handler.Server{}
	got := h.Start(t.Context(), testLog, nil, config.Config{})
	be.True(t, errors.Is(got, handler.ErrNoPort))
}

func TestRender(t *testing.T) {
	t.Parallel()

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		tr := new(handler.TemplateRegistry)
		c := echo.New().NewContext(nil, nil)

		be.Err(t, tr.Render(nil, nil, "", nil))
		be.Err(t, tr.Render(c, io.Discard, "name", nil))
		be.True(t, errors.Is(tr.Render(c, io.Discard, "", "data"), handler.ErrNoName))
		be.True(t, errors.Is(tr.Render(c, io.Discard, "missing", "data"), handler.ErrNoTmpl))
	})

	t.Run("layout", func(t *testing.T) {
		t.Parallel()

		tmpl := template.Must(template.New("layout").Parse("hello {{.}}"))
		tr := &handler.TemplateRegistry{
			Templates: map[string]*template.Template{"ok": tmpl},
		}
		var buf bytes.Buffer
		c := echo.New().NewContext(nil, nil)
		be.Err(t, tr.Render(c, &buf, "ok", "world"), nil)
		be.Equal(t, buf.String(), "hello world")
	})
}

func TestCerts(t *testing.T) {
	t.Parallel()

	root := testutil.OpenRoot(t)
	h := handler.Server{
		Public: root.FS(),
		Environment: config.Config{
			TLSPort: 1234,
		},
	}

	_, _, _, err := h.Local(t.Context(), nil)
	be.Err(t, err)

	x, cert, key, err := h.Local(t.Context(), testLog)
	be.Err(t, err, nil)
	be.Equal(t, x.Address, ":1234")

	_, err = tls.X509KeyPair(cert, key)
	be.Err(t, err, nil)
}

func TestPrint(t *testing.T) {
	t.Parallel()

	t.Run("offline", func(t *testing.T) {
		t.Parallel()

		c := handler.Server{}
		be.Err(t, c.Print(nil), nil)

		var w bytes.Buffer
		be.Err(t, c.Print(&w), nil)
		out := w.String()
		be.True(t, strings.Contains(out, "SYSTEM"))
		be.True(t, strings.Contains(out, "STATUS"))
		be.True(t, strings.Contains(out, "offline"))
		be.True(t, strings.Contains(out, "none"))
	})

	t.Run("configured", func(t *testing.T) {
		t.Parallel()

		c := handler.Server{
			Version:     "v1.2.3",
			DBVersion:   "PostgreSQL 17",
			RecordCount: 66082,
			Environment: config.Config{
				HTTPPort:    config.StdCustom,
				TLSPort:     config.PortTLS(8443),
				MatchHost:   config.Matchhost("127.0.0.1"),
				ProdMode:    true,
				NoCrawl:     true,
				LogAll:      true,
				DemozooAuth: config.DemozooAuth("badkey"),
			},
		}

		var w bytes.Buffer
		be.Err(t, c.Print(&w), nil)
		out := w.String()
		be.True(t, strings.Contains(out, "v1.2.3"))
		be.True(t, strings.Contains(out, "PostgreSQL 17 · 66,082 artifacts"))
		be.True(t, strings.Contains(out, "production · write mode · no-crawl · gzip off · log-all"))
		be.True(t, strings.Contains(out, "HTTP :1323 · HTTPS :8443 (self-signed) (host 127.0.0.1)"))
		be.True(t, strings.Contains(out, "CONFIG ISSUES"))
		be.True(t, strings.Contains(out, "D2_DEMOZOO_KEY"))
		be.True(t, !strings.Contains(out, "badkey"))
	})

	t.Run("below minimum", func(t *testing.T) {
		t.Parallel()

		c := handler.Server{
			DBVersion:   "PostgreSQL 17",
			RecordCount: 100,
		}
		var w bytes.Buffer
		be.Err(t, c.Print(&w), nil)
		be.True(t, strings.Contains(w.String(), "100 artifacts (below minimum)"))
	})
}

func TestHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  config.Config
		want string
	}{
		{"disabled", config.Config{}, ""},
		{"port", config.Config{HTTPPort: config.StdCustom}, ":1323"},
		{"match host", config.Config{HTTPPort: config.StdCustom, MatchHost: "127.0.0.1"}, "127.0.0.1:1323"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.Server{Environment: tt.env}
			got := h.HTTP()
			be.Equal(t, got.Address, tt.want)
			be.True(t, got.HideBanner)
			be.True(t, got.HidePort)
		})
	}
}

func TestTLSConfig(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		h := handler.Server{}
		cfg, cert, key, err := h.TLS(t.Context(), testLog)
		be.Err(t, err, nil)
		be.Equal(t, cfg.Address, "")
		be.Equal(t, len(cert), 0)
		be.Equal(t, len(key), 0)
	})

	t.Run("missing paths", func(t *testing.T) {
		t.Parallel()

		h := handler.Server{Environment: config.Config{TLSPort: 8443}}
		_, _, _, err := h.TLS(t.Context(), testLog)
		be.True(t, errors.Is(err, handler.ErrNoTLS))
	})

	t.Run("missing files", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		h := handler.Server{
			Environment: config.Config{
				TLSPort: 8443,
				TLSCert: config.Abstlscrt(filepath.Join(tmp, "no-cert.pem")),
				TLSKey:  config.Abstlskey(filepath.Join(tmp, "no-key.pem")),
			},
		}
		_, _, _, err := h.TLS(t.Context(), testLog)
		be.True(t, errors.Is(err, handler.ErrNoTLS))
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		certPath := filepath.Join(tmp, "cert.pem")
		keyPath := filepath.Join(tmp, "key.pem")
		be.Err(t, os.WriteFile(certPath, []byte("CERTDATA"), 0o600), nil)
		be.Err(t, os.WriteFile(keyPath, []byte("KEYDATA"), 0o600), nil)

		h := handler.Server{
			Environment: config.Config{
				TLSPort: 8443,
				TLSCert: config.Abstlscrt(certPath),
				TLSKey:  config.Abstlskey(keyPath),
			},
		}
		cfg, cert, key, err := h.TLS(t.Context(), testLog)
		be.Err(t, err, nil)
		be.Equal(t, cfg.Address, ":8443")
		be.Equal(t, string(cert), "CERTDATA")
		be.Equal(t, string(key), "KEYDATA")
	})
}

// TestPrintRoutes mutates package-level app.Caching via registerRoutes -> RouteFS, so it cannot run in parallel.
func TestPrintRoutes(t *testing.T) { //nolint:paralleltest
	root := testutil.OpenRoot(t)

	t.Run("all", func(t *testing.T) {
		h := handler.Server{Public: root.FS()}
		var buf bytes.Buffer
		be.Err(t, h.PrintRoutes(&buf, ""), nil)
		out := buf.String()
		be.True(t, strings.Contains(out, "Web Server Routes ("))
		be.True(t, strings.Contains(out, "/health-check"))
		be.True(t, !strings.Contains(out, "\033["))
	})

	t.Run("method filter", func(t *testing.T) {
		h := handler.Server{Public: root.FS()}
		var buf bytes.Buffer
		be.Err(t, h.PrintRoutes(&buf, "post"), nil)
		out := buf.String()
		be.True(t, strings.Contains(out, "POST)"))
		be.True(t, strings.Contains(out, "  POST"))
		be.True(t, !strings.Contains(out, "  GET "))
	})

	t.Run("pprof in dev log-all", func(t *testing.T) {
		h := handler.Server{
			Public: root.FS(),
			Environment: config.Config{
				ProdMode: false,
				LogAll:   true,
			},
		}
		var buf bytes.Buffer
		be.Err(t, h.PrintRoutes(&buf, ""), nil)
		be.True(t, strings.Contains(buf.String(), "/debug/pprof"))
	})
}

func TestRegistry(t *testing.T) {
	t.Parallel()

	c := handler.Server{}
	x, err := c.TemplRegistry(t.Context(), nil, nil)
	be.Err(t, err)
	be.True(t, x == nil)
}

func TestRouteFS(t *testing.T) {
	t.Parallel()

	root := testutil.OpenRoot(t)
	e := handler.RouteFS(echo.New(), root.FS())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/image/new", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	be.Equal(t, rec.Code, http.StatusNotFound)

	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/image/new/navbar.webp", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	be.Equal(t, rec.Code, http.StatusOK)
}

// TestIndexRender mutates package-level app.Caching via Handler -> RouteFS, so it cannot run in parallel.
func TestIndexRender(t *testing.T) { //nolint:paralleltest
	fsys := testutil.ProjectFS(t)
	db, err := sql.Open("pgx", "")
	be.Err(t, err, nil)
	t.Cleanup(func() {
		_ = db.Close()
	})

	h := handler.Server{
		Environment: config.Config{
			ReadOnly: true,
			Quiet:    true,
		},
		Public:      fsys,
		View:        fsys,
		RecordCount: 66082,
		TotalBytes:  21_000_000_000,
		MinYear:     1981,
		MaxYear:     2026,
	}

	e := h.Handler(t.Context(), testLog, db)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	be.Equal(t, rec.Code, http.StatusOK)
	body := rec.Body.String()
	be.Equal(t, strings.Count(body, `id="milestoneModal"`), 1)
	be.Equal(t, strings.Count(body, `id=""`), 0)
	be.True(t, strings.Contains(body, `66,082 artifacts · 21.0 GB · 1981–2026`))
	be.True(t, strings.Contains(body, `<meta property="og:description"`))
	be.True(t, strings.Contains(body, `/js/index.min.js`))
	be.True(t, !strings.Contains(body, `<style scoped>`))
}
