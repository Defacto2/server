// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
package handler

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/out"
	"github.com/Defacto2/server/internal/panics"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// ShutdownCounter is the number of iterations to wait before shutting down the server.
	ShutdownCounter = 3

	// ShutdownWait is the number of seconds to wait before shutting down the server.
	ShutdownWait = ShutdownCounter * time.Second

	// Downloader is the route for the file download handler.
	Downloader = "/d/:id"
)

var (
	ErrNoName = errors.New("name is empty")
	ErrNoTmpl = errors.New("named template cannot be found")
	ErrNoPort = errors.New("web server ports are not configured")
)

// Configuration of the handler.
type Configuration struct {
	Public      embed.FS      // Public facing files.
	View        embed.FS      // View contains Go templates.
	Version     string        // Version is the results of GoReleaser build command.
	Brand       []byte        // Brand contains the Defacto2 ASCII logo.
	Environment config.Config // Environment configurations from the host system.
	RecordCount int           // The total number of file records in the database.
}

// Controller is the primary instance of the Echo router.
func (c *Configuration) Controller(db *sql.DB, sl *slog.Logger) *echo.Echo {
	const msg = "controller handler"
	configs := c.Environment
	e := echo.New()
	if configs.LogAll {
		pprof.Register(e) // TODO: test the logall flag
	}
	e.HideBanner = true
	// TODO: test
	customerrorhandler := func(err error, ctx echo.Context) {
		configs.CustomErrorHandler(err, ctx, sl)
	}
	e.HTTPErrorHandler = customerrorhandler
	// configs.CustomErrorHandler
	// e.HTTPErrorHandler = configs.CustomErrorHandler

	tmpl, err := c.Registry(db, sl)
	if err != nil {
		out.Fatal(sl, msg, slog.String("template", "could not register the templates"), slog.Any("fatal", err))
	}
	e.Renderer = tmpl
	middlewares := []echo.MiddlewareFunc{
		middleware.Rewrite(rewrites()),
		middleware.NonWWWRedirect(),
	}
	e.Pre(middlewares...)

	// *************************************************
	//  NOTE: NEVER USE the middleware.Timeout()
	//   It is broken and should not be in the
	//   labstack/echo library, as it can easily crash!
	// *************************************************
	middlewares = []echo.MiddlewareFunc{
		middleware.Secure(),
		middleware.RequestLoggerWithConfig(c.configSlog()),
		c.NoCrawl,
		middleware.RemoveTrailingSlashWithConfig(configRTS()),
	}
	if configs.Compression {
		middlewares = append(middlewares, middleware.Gzip())
	}
	if configs.ProdMode {
		middlewares = append(middlewares, middleware.Recover())
	}
	e.Use(middlewares...)
	e = EmbedDirs(e, c.Public)
	e = MovedPermanently(e)
	e = htmxGroup(e, db, sl, bool(configs.ProdMode), dir.Directory(c.Environment.AbsDownload))
	e, err = c.FilesRoutes(e, db, sl, c.Public)
	if err != nil {
		out.Fatal(sl, msg, slog.String("file routes", "could not register the routes"), slog.Any("fatal", err))
	}
	group := html3.Routes(e, db, sl)
	group.GET(Downloader, func(cx echo.Context) error {
		return c.downloader(cx, db, sl)
	})
	return e
}

// EmbedDirs serves the static files from the directories embed to the binary.
func EmbedDirs(e *echo.Echo, currentFs fs.FS) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for the embed directories binary", panics.ErrNoEchoE))
	}
	dirs := map[string]string{
		"/image/artpack":   "public/image/artpack",
		"/image/html3":     "public/image/html3",
		"/image/layout":    "public/image/layout",
		"/image/milestone": "public/image/milestone",
		"/image/new":       "public/image/new",
		"/svg":             "public/svg",
		"/jsdos/bin":       "public/bin/dos32",
	}
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(currentFs, fsRoot))
		e.GET(path, func(_ echo.Context) error {
			return echo.NewHTTPError(http.StatusNotFound)
		})
	}
	return e
}

// Info prints the application information to the console.
func (c *Configuration) Info(sl *slog.Logger, w io.Writer) {
	const msg = "configuration info"
	if w == nil {
		w = io.Discard
	}
	nr := bytes.NewReader(c.Brand)
	if l, err := io.Copy(w, nr); err != nil {
		sl.Warn(msg, slog.String("brand", "could not print the startup logo"), slog.Any("error", err))
	} else if l > 0 {
		_, err := fmt.Fprint(w, "\n\n")
		if err != nil {
			panic(err)
		}
	}

	_, err := fmt.Fprintf(w, "  %s.\n", flags.Copyright())
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(w, "%s\n", c.versionBrief())

	cpuInfo := fmt.Sprintf("  %d active routines sharing %d usable threads on %d CPU cores.",
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	_, _ = fmt.Fprintln(w, cpuInfo)

	golangInfo := fmt.Sprintf("  Compiled on Go %s for %s with %s.\n",
		runtime.Version()[2:], flags.OS(), flags.Arch())
	_, _ = fmt.Fprintln(w, golangInfo)
	//
	// All additional feedback should go in internal/config/check.go (c *Config) Checks()
	//
}

// PortErr handles the error when the HTTP or HTTPS server cannot start.
func (c *Configuration) PortErr(sl *slog.Logger, port uint, err error) {
	const msg = "http/https"
	s := "HTTP"
	if port == c.Environment.TLSPort.Value() {
		s = "TLS"
	}
	var portErr *net.OpError
	switch {
	case !bool(c.Environment.ProdMode) && errors.As(err, &portErr):
		sl.Warn("air or task",
			slog.String("problem", "could not startup air or task"),
			slog.String("help", "however, this issue can probably be ignored"),
			slog.Any("error", err))
	case errors.Is(err, net.ErrClosed),
		errors.Is(err, http.ErrServerClosed):
		sl.Info("shutdown",
			slog.String("success", fmt.Sprintf("the %s server will gracefully shutdown", s)))
	case errors.Is(err, os.ErrNotExist):
		out.Fatal(sl, msg,
			slog.String("port error", "could not startup the server using the configured port"),
			slog.Int("port", int(port)), slog.Any("error", err))
	default:
	}
}

// Registry returns the template renderer.
func (c *Configuration) Registry(db *sql.DB, sl *slog.Logger) (*TemplateRegistry, error) {
	webapp := app.Templ{
		Environment: c.Environment,
		Brand:       c.Brand,
		Public:      c.Public,
		RecordCount: c.RecordCount,
		Version:     c.Version,
		View:        c.View,
	}
	tmpls, err := webapp.Templates(db)
	if err != nil {
		return nil, fmt.Errorf("handler registry, %w", err)
	}
	src := html3.Templates(db, sl, c.View)
	maps.Copy(tmpls, src)
	src = htmx.Templates(c.View)
	maps.Copy(tmpls, src)
	return &TemplateRegistry{Templates: tmpls}, nil
}

// ShutdownHTTP waits for a Ctrl-C keyboard press to initiate a graceful shutdown of the HTTP web server.
// The shutdown procedure occurs a few seconds after the key press.
func (c *Configuration) ShutdownHTTP(e *echo.Echo, sl *slog.Logger) {
	const msg = "shutdown"
	if e == nil {
		panic(fmt.Errorf("%w for the HTTP shutdown", panics.ErrNoEchoE))
	}
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	waitDuration := ShutdownWait
	waitCount := ShutdownCounter
	ticker := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer func() {
		const alert = "Detected Ctrl + C, server will shutdown"
		//_ = logger.Sync() // do not check Sync errors as there can be false positives
		w := os.Stdout
		buf := bufio.NewWriter(w)
		_, err := fmt.Fprintf(buf, "\n%s in %v ", alert, waitDuration)
		if err != nil {
			panic(err)
		}
		err = buf.Flush()
		if err != nil {
			panic(err)
		}
		count := waitCount
		pause := time.NewTicker(ticker)
		for range pause.C {
			count--
			buf := bufio.NewWriter(w)
			if count <= 0 {
				_, err := fmt.Fprintf(buf, "\r%s %s\n", alert, "now     ")
				if err != nil {
					panic(err)
				}
				err = buf.Flush()
				if err != nil {
					panic(err)
				}
				break
			}
			_, err = fmt.Fprintf(buf, "\r%s in %ds ", alert, count)
			if err != nil {
				panic(err)
			}
			err = buf.Flush()
			if err != nil {
				panic(err)
			}
		}
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
		if err := e.Shutdown(ctx); err != nil {
			out.Fatal(sl, msg, slog.String("context", "caused an error"), slog.Any("error", err))
		}
		sl.Info(msg, slog.String("success", "shutdown complete"))
		signal.Stop(quit)
		cancel()
	}()
}

// Start the HTTP, and-or the TLS servers.
func (c *Configuration) Start(e *echo.Echo, sl *slog.Logger, configs config.Config) error {
	if e == nil {
		panic(fmt.Errorf("%w for the web application startup", panics.ErrNoEchoE))
	}
	switch {
	case configs.UseTLS() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			c.StartHTTP(e2, sl)
		}()
		go c.StartTLS(e, sl)
	case configs.UseTLSLocal() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			c.StartHTTP(e2, sl)
		}()
		go c.StartTLSLocal(e, sl)
	case configs.UseTLS():
		go c.StartTLS(e, sl)
	case configs.UseHTTP():
		go c.StartHTTP(e, sl)
	case configs.UseTLSLocal():
		go c.StartTLSLocal(e, sl)
	default:
		return ErrNoPort
	}
	return nil
}

// StartHTTP starts the insecure HTTP web server.
func (c *Configuration) StartHTTP(e *echo.Echo, sl *slog.Logger) {
	if e == nil {
		panic(fmt.Errorf("%w for the HTTP startup", panics.ErrNoEchoE))
	}
	port := c.Environment.HTTPPort.Value()
	address := c.address(port)
	if address == "" {
		return
	}
	if err := e.Start(address); err != nil {
		c.PortErr(sl, port, err)
	}
}

// StartTLS starts the encrypted TLS web server.
func (c *Configuration) StartTLS(e *echo.Echo, sl *slog.Logger) {
	const msg = "tls web server"
	if e == nil {
		panic(fmt.Errorf("%w for the TLS startup", panics.ErrNoEchoE))
	}
	port := c.Environment.TLSPort.Value()
	address := c.address(port)
	if address == "" {
		return
	}
	certFile := c.Environment.TLSCert
	keyFile := c.Environment.TLSKey
	if certFile == "" || keyFile == "" {
		out.Fatal(sl, msg,
			slog.String("failure", "missing critical file"),
			slog.String("certificate file", string(certFile)),
			slog.String("key file", string(keyFile)))
	}
	if !helper.File(certFile.String()) {
		out.Fatal(sl, msg,
			slog.String("certificate file", "file does not exist"))
	}
	if !helper.File(keyFile.String()) {
		out.Fatal(sl, msg,
			slog.String("key file", "file does not exist"))
	}
	if err := e.StartTLS(address, certFile, keyFile); err != nil {
		c.PortErr(sl, port, err)
	}
}

// StartTLSLocal starts the localhost, encrypted TLS web server.
// This should only be triggered when the server is running in local mode.
func (c *Configuration) StartTLSLocal(e *echo.Echo, sl *slog.Logger) {
	const msg = "tls localhost server"
	if e == nil {
		panic(fmt.Errorf("%w for the TLS local mode startup", panics.ErrNoEchoE))
	}
	port := c.Environment.TLSPort.Value()
	address := c.address(port)
	if address == "" {
		return
	}
	const cert, key = "public/certs/cert.pem", "public/certs/key.pem"
	certB, err := c.Public.ReadFile(cert)
	if err != nil {
		out.Fatal(sl, msg,
			slog.String("certificate", "read file failure"), slog.Any("error", err))
	}
	keyB, err := c.Public.ReadFile(key)
	if err != nil {
		out.Fatal(sl, msg,
			slog.String("key", "read file failure"), slog.Any("error", err))
	}
	if err := e.StartTLS(address, certB, keyB); err != nil {
		c.PortErr(sl, port, err)
	}
}

func (c *Configuration) address(port uint) string {
	if port == 0 {
		return ""
	}
	address := fmt.Sprintf(":%d", port)
	if c.Environment.MatchHost != "" {
		address = fmt.Sprintf("%s:%d", c.Environment.MatchHost, port)
	}
	return address
}

// downloader route for the file download handler under the html3 group.
func (c *Configuration) downloader(cx echo.Context, db *sql.DB, sl *slog.Logger) error {
	d := download.Download{
		Inline: false,
		Dir:    dir.Directory(c.Environment.AbsDownload),
	}
	if err := d.HTTPSend(cx, db, sl); err != nil {
		return fmt.Errorf("d.HTTPSend: %w", err)
	}
	return nil
}

// versionBrief returns the application version string.
func (c *Configuration) versionBrief() string {
	if c.Version == "" {
		return "  no version info, app compiled binary directly."
	}
	return fmt.Sprintf("  %s.", flags.Commit(c.Version))
}

// Rewrites for assets.
// This is different to a redirect as it keeps the original URL in the browser.
func rewrites() map[string]string {
	return map[string]string{
		"/logo.txt": "/text/defacto2.txt",
	}
}

// TemplateRegistry is template registry struct.
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

// Render the layout template with the core HTML, META and BODY elements.
func (t *TemplateRegistry) Render(w io.Writer, name string, data any, c echo.Context) error {
	const layout, info = "layout", "template registry render"
	if name == "" {
		return fmt.Errorf("%s layout: %w", info, ErrNoName)
	}
	if w == nil {
		return fmt.Errorf("%s io.writer is nil: %w", info, echo.ErrRendererNotRegistered)
	}
	if data == nil {
		return fmt.Errorf("%s data interface is nil: %w", info, echo.ErrRendererNotRegistered)
	}
	if c == nil {
		return fmt.Errorf("%s echo context is nil: %w", info, echo.ErrRendererNotRegistered)
	}
	tmpl, exists := t.Templates[name]
	if !exists {
		return fmt.Errorf("registry render %w: %q", ErrNoTmpl, name)
	}
	if err := tmpl.ExecuteTemplate(w, layout, data); err != nil {
		return fmt.Errorf("%s execute: %w", info, err)
	}
	return nil
}
