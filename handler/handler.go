// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
package handler

import (
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
	"runtime"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/panics"
	"github.com/labstack/echo-contrib/v5/pprof"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"golang.org/x/sync/errgroup"
)

const (
	// ShutdownCounter is the number of iterations to wait before shutting down the server.
	ShutdownCounter = 1

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
	Public      embed.FS         // Public facing files.
	View        embed.FS         // View contains Go templates.
	Version     string           // Version is the results of GoReleaser build command.
	Brand       []byte           // Brand contains the Defacto2 ASCII logo.
	Environment config.Config    // Environment configurations from the host system.
	RecordCount int              // The total number of file records in the database.
	TidbitIndex fulltext.Tidbits // Fulltext search index of the tidbit markdown files.
}

// Handler is the primary instance of the Echo router.
func (c *Configuration) Handler(sl *slog.Logger, db *sql.DB) *echo.Echo {
	const msg = "controller handler"
	err := panics.SD(sl, db)
	if err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	envConfig := c.Environment
	config := echo.Config{
		Logger: sl,
		HTTPErrorHandler: func(ctx *echo.Context, err error) {
			envConfig.CustomErrorHandler(err, ctx, sl)
		},
		// Router:             nil, // TODO
		// OnAddRoute:         nil,
		// Filesystem:         nil,
		// Binder:             nil,
		// Validator:          nil,
		// Renderer:           nil,
		// JSONSerializer:     nil,
		// IPExtractor:        nil,
		// FormParseMaxMemory: 0, // TODO
	}

	config.Renderer, err = c.TemplRegistry(db, sl)
	if err != nil {
		logs.Fatal(sl, msg,
			slog.String("template", "could not register the templates"),
			slog.Any("fatal", err))
	}

	e := echo.NewWithConfig(config)
	if envConfig.LogAll {
		// echo prefix options that get used by RequestLoggerConfig
		pprof.Register(e)
	}
	// pre middleware
	mid := []echo.MiddlewareFunc{}
	mid = append(mid, middleware.Rewrite(rewrites()))
	mid = append(mid, middleware.NonWWWRedirect())
	e.Pre(mid...)
	// use middleware
	mid = []echo.MiddlewareFunc{}
	mid = append(mid, middleware.Secure())
	mid = append(mid, middleware.RequestLoggerWithConfig(c.RequestLoggerConfig(sl)))
	mid = append(mid, c.NoCrawl)
	mid = append(mid, middleware.RemoveTrailingSlashWithConfig(configTrailSlash()))
	if envConfig.Compression {
		mid = append(mid, middleware.Gzip())
	}
	if envConfig.ProdMode {
		mid = append(mid, middleware.Recover())
	}
	e.Use(mid...)

	e = EmbedDirs(e, c.Public)
	e = MovedPermanently(e)
	e = htmxGroup(e, db, sl, bool(envConfig.ProdMode), dir.Directory(c.Environment.AbsDownload))
	e, err = c.FilesRoutes(e, db, sl, c.Public)
	if err != nil {
		logs.Fatal(sl, msg,
			slog.String("file routes", "could not register the routes"),
			slog.Any("fatal", err))
	}
	group := html3.Routes(e, db, sl)
	group.GET(Downloader, func(cx *echo.Context) error {
		return c.downloader(cx, db, sl)
	})
	return e
}

// EmbedDirs serves the static files from the directories embed to the binary.
func EmbedDirs(e *echo.Echo, currentFs fs.FS) *echo.Echo {
	const msg = "embed dirs handler"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	dirs := map[string]string{
		"/image/artpack":   "public/image/artpack",
		"/image/html3":     "public/image/html3",
		"/image/layout":    "public/image/layout",
		"/image/milestone": "public/image/milestone",
		"/image/new":       "public/image/new",
		"/svg":             "public/svg",
		"/jsdos/bin":       "public/bin/dos32",
		"/js":              "public/js",
	}
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(currentFs, fsRoot))
		// Block directory listing; allows files to be served but returns 404 for directory itself
		e.GET(path, func(_ *echo.Context) error {
			return echo.NewHTTPError(http.StatusNotFound, "directory not found")
		})
	}
	return e
}

// Print the application logo and information to the w io.writer.
func (c *Configuration) Print(sl *slog.Logger, w io.Writer) {
	const msg = "configuration info handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	if w == nil {
		w = io.Discard
	}
	nr := bytes.NewReader(c.Brand)
	if l, err := io.Copy(w, nr); err != nil {
		sl.Warn(msg, slog.String("brand", "could not print the startup logo"), slog.Any("error", err))
	} else if l > 0 {
		_, err := fmt.Fprint(w, "\n\n")
		if err != nil {
			panic(fmt.Errorf("%s: %w", msg, err))
		}
	}
	_, err := fmt.Fprintf(w, "  %s.\n", flags.Copyright())
	if err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	_, _ = fmt.Fprintf(w, "%s\n", c.versionBrief())
	cpuInfo := fmt.Sprintf("  %d active routines sharing %d usable threads on %d CPU cores.",
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	_, _ = fmt.Fprintln(w, cpuInfo)
	golangInfo := fmt.Sprintf("  Compiled on Go %s for %s with %s.",
		runtime.Version()[2:], flags.OS(), flags.Arch())
	_, _ = fmt.Fprintln(w, golangInfo)
	to, fr, _, s, err := helper.DiskStat("/")
	if err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	total := int64(to)
	free := int64(fr)
	// Disk (/): 300.28 GiB / 464.17 GiB (65%)
	diskInfo := fmt.Sprintf("  Disk (/) %s / %s (%s).\n",
		helper.ByteCount(free), helper.ByteCount(total),
		s)
	_, _ = fmt.Fprintln(w, diskInfo)
	//
	// All additional feedback should go in internal/config/check.go (c *Config) Checks()
	//
}

// PortErr handles the error when the HTTP or HTTPS server cannot start.
func (c *Configuration) PortErr(sl *slog.Logger, port uint16, err error) {
	const msg = "http/https"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
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
		logs.Fatal(sl, msg,
			slog.String("port error", "could not startup the server using the configured port"),
			slog.Int("port", int(port)), slog.Any("error", err))
	default:
		sl.Warn(s+" server startup failed",
			slog.Any("error", err))
	}
}

// TemplRegistry returns the template registry for the renderer.
func (c *Configuration) TemplRegistry(db *sql.DB, sl *slog.Logger) (*TemplateRegistry, error) {
	const msg = "template registry handler"
	if err := panics.SD(sl, db); err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	webapp := app.Templ{
		Public:      c.Public,
		View:        c.View,
		Subresource: app.SRI{}, //nolint:exhaustruct // SRI fields are computed via Verify() method
		Version:     c.Version,
		Brand:       c.Brand,
		Environment: c.Environment,
		RecordCount: c.RecordCount,
	}
	tmpls, err := webapp.Templates(db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	src := html3.Templates(db, sl, c.View)
	maps.Copy(tmpls, src)
	src = htmx.Templates(c.View)
	maps.Copy(tmpls, src)
	return &TemplateRegistry{Templates: tmpls}, nil
}

// ShutdownHTTP waits for a Ctrl-C keyboard press to initiate a graceful shutdown of the HTTP web server.
// The shutdown procedure occurs a few seconds after the key press.
// TODO: retire
// func (c *Configuration) ShutdownHTTP(w io.Writer, e *echo.Echo, sl *slog.Logger) { //nolint:funlen
// 	if w == nil {
// 		w = os.Stderr
// 	}
// 	const msg = "shutdown http handler"
// 	if err := panics.EchoS(e, sl); err != nil {
// 		panic(fmt.Errorf("%s: %w", msg, err))
// 	}
// 	// Wait for interrupt signal to gracefully shutdown the server
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, os.Interrupt)
// 	<-quit
// 	waitDuration := ShutdownWait
// 	waitCount := ShutdownCounter
// 	ticker := 1 * time.Second
// 	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
// 	defer cancel()
// 	defer func() {
// 		const alert = "Detected Ctrl + C, server will shutdown"
// 		// _ = logger.Sync() // do not check Sync errors as there can be false positives
// 		buf := bufio.NewWriter(w)
// 		_, err := fmt.Fprintf(buf, "\n%s in %v ", alert, waitDuration)
// 		if err != nil {
// 			panic(err)
// 		}
// 		err = buf.Flush()
// 		if err != nil {
// 			panic(err)
// 		}
// 		count := waitCount
// 		pause := time.NewTicker(ticker)
// 		for range pause.C {
// 			count--
// 			buf.Reset(w)
// 			if count <= 0 {
// 				_, err := fmt.Fprintf(buf, "\r%s %s\n", alert, "now     ")
// 				if err != nil {
// 					panic(err)
// 				}
// 				err = buf.Flush()
// 				if err != nil {
// 					panic(err)
// 				}
// 				pause.Stop()
// 				break
// 			}
// 			_, err = fmt.Fprintf(buf, "\r%s in %ds ", alert, count)
// 			if err != nil {
// 				panic(err)
// 			}
// 			err = buf.Flush()
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 		select {
// 		case <-quit:
// 			cancel()
// 		case <-ctx.Done():
// 		}
// 		const shutdownTimeout = 5 * time.Second
// 		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
// 		defer shutdownCancel()
// 		if err := e.Shutdown(shutdownCtx); err != nil {
// 			logs.FatalTx(shutdownCtx, sl, msg,
// 				slog.String("context", "caused an error"), slog.Any("error", err))
// 		}
// 		sl.Info(msg, slog.String("success", "shutdown complete"))
// 		signal.Stop(quit)
// 		cancel()
// 	}()
// }

func (c *Configuration) EchoConfig() echo.StartConfig {
	config := echo.StartConfig{
		Address:    "",
		HideBanner: true,
		HidePort:   false,
		// CertFilesystem:   nil,
		// TLSConfig:        nil,
		// Listener:         nil,
		// ListenerNetwork:  "",
		// ListenerAddrFunc: nil,
		GracefulTimeout: time.Duration(3 * time.Second),
		// OnShutdownError:  nil,
		// BeforeServeFunc:  nil,
	}
	return config
}

// Start the HTTP, and-or the TLS servers.
func (c *Configuration) Start(ctx context.Context, sl *slog.Logger, h http.Handler, configs config.Config) error {
	const msg = "start server handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	switch {
	case configs.UseTLS() && configs.UseHTTP():
		c.StartDual(ctx, sl, h)
	case configs.UseLocal() && configs.UseHTTP():
		c.StartLocals(ctx, sl, h)
	case configs.UseTLS():
		c.StartTLS(ctx, sl, h)
	case configs.UseHTTP():
		c.StartHTTP(ctx, sl, h)
	case configs.UseLocal():
		c.StartLocal(ctx, sl, h)
	default:
		return fmt.Errorf("%s: %w", msg, ErrNoPort)
	}
	return nil
}

func (c *Configuration) StartLocals(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start locals handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	c.startDual(ctx, sl, h, true)
}

func (c *Configuration) StartDual(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start duals handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	c.startDual(ctx, sl, h, false)
}

func (c *Configuration) startDual(ctx context.Context, sl *slog.Logger, h http.Handler, local bool) {
	const msg = "start dual handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	g, ctx := errgroup.WithContext(ctx)

	httpConfig := c.HTTP()
	httpsConfig := echo.StartConfig{}
	certFile, keyFile := "", ""
	if local {
		httpsConfig, certFile, keyFile = c.Local(sl)
	} else {
		httpsConfig, certFile, keyFile = c.TLS(sl)
	}

	g.Go(func() error {
		sl.Info("Starting HTTP Listener", "address", httpConfig.Address)
		err := httpConfig.Start(ctx, h)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sl.Error("HTTP Server crashed unexpectedly", "error", err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		sl.Info("Starting HTTPS Listener", "address", httpsConfig.Address)
		// Point to your valid SSL/TLS files
		err := httpsConfig.StartTLS(ctx, h, certFile, keyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sl.Error("HTTPS Server crashed unexpectedly", "error", err)
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		sl.Error("System tracking intercepted a service failure", "error", err)
		return
	}

	sl.Info("Dual server infrastructure successfully stopped.")
}

func (c *Configuration) HTTP() echo.StartConfig {
	config := c.EchoConfig()
	port := c.Environment.HTTPPort.Value()
	address := c.address(port)
	if address == "" {
		return config
	}
	config.Address = address
	return config
}

func (c *Configuration) Local(sl *slog.Logger) (echo.StartConfig, string, string) {
	const msg = "start local tls handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	config := c.EchoConfig()
	port := c.Environment.TLSPort.Value()
	address := c.address(port)
	if address == "" {
		return config, "", ""
	}
	config.Address = address

	certFile, err := c.Public.ReadFile("public/certs/cert.pem")
	if err != nil {
		logs.Fatal(sl, msg,
			slog.String("certificate", "read file failure"), slog.Any("error", err))
	}
	keyFile, err := c.Public.ReadFile("public/certs/key.pem")
	if err != nil {
		logs.Fatal(sl, msg,
			slog.String("key", "read file failure"), slog.Any("error", err))
	}
	return config, string(certFile), string(keyFile)
}

func (c *Configuration) TLS(sl *slog.Logger) (echo.StartConfig, string, string) {
	const msg = "start tls handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	config := c.EchoConfig()
	port := c.Environment.TLSPort.Value()
	address := c.address(port)
	if address == "" {
		return config, "", ""
	}
	config.Address = address
	certFile := c.Environment.TLSCert
	keyFile := c.Environment.TLSKey
	if certFile == "" || keyFile == "" {
		logs.Fatal(sl, msg,
			slog.String("failure", "missing critical file"),
			slog.String("certificate file", string(certFile)),
			slog.String("key file", string(keyFile)))
	}
	if !helper.File(certFile.String()) {
		logs.Fatal(sl, msg,
			slog.String("certificate file", "file does not exist"))
	}
	if !helper.File(keyFile.String()) {
		logs.Fatal(sl, msg,
			slog.String("key file", "file does not exist"))
	}
	return config, certFile.String(), keyFile.String()
}

// StartHTTP starts the insecure HTTP web server.
func (c *Configuration) StartHTTP(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start http handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	httpConfig := c.HTTP()
	sl.Info("Starting HTTP Listener", "address", httpConfig.Address)
	err := httpConfig.Start(ctx, h)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTP Server crashed unexpectedly", "error", err)
		return
	}
}

// StartTLS starts the encrypted TLS web server.
func (c *Configuration) StartTLS(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start tls handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	httpsConfig, certFile, keyFile := c.TLS(sl)
	sl.Info("Starting HTTPS Listener", "address", httpsConfig.Address)
	// Point to your valid SSL/TLS files
	err := httpsConfig.StartTLS(ctx, h, certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTPS Server crashed unexpectedly", "error", err)
		return
	}
}

// StartTLSLocal starts the insecure localhost, encrypted TLS web server.
// This should only be triggered when the server is running in local mode.
func (c *Configuration) StartLocal(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start local tls handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	httpsConfig, certFile, keyFile := c.Local(sl)
	sl.Info("Starting HTTPS Listener", "address", httpsConfig.Address)
	// Point to your valid SSL/TLS files
	err := httpsConfig.StartTLS(ctx, h, certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTPS Server crashed unexpectedly", "error", err)
		return
	}
}

func (c *Configuration) address(port uint16) string {
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
func (c *Configuration) downloader(ctx *echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "downloader htm3 group handler"
	if err := panics.EchoContextDS(ctx, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	d := download.Download{
		Inline: false,
		Dir:    dir.Directory(c.Environment.AbsDownload),
	}
	if err := d.HTTPSend(ctx, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
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
func (t *TemplateRegistry) Render(c *echo.Context, w io.Writer, name string, data any) error {
	const msg = "template registry render handler"
	if name == "" {
		return fmt.Errorf("%s name layout: %w", msg, ErrNoName)
	}
	if w == nil {
		return fmt.Errorf("%s w io.writer is nil: %w", msg, echo.ErrRendererNotRegistered)
	}
	if data == nil {
		return fmt.Errorf("%s data interface is nil: %w", msg, echo.ErrRendererNotRegistered)
	}
	if c == nil {
		return fmt.Errorf("%s c echo context is nil: %w", msg, echo.ErrRendererNotRegistered)
	}
	tmpl, exists := t.Templates[name]
	if !exists {
		return fmt.Errorf("%s %q: %w", msg, name, ErrNoTmpl)
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		return fmt.Errorf("%s execute of '%s': %w", msg, name, err)
	}
	return nil
}
