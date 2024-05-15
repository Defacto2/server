// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
package handler

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/handler/middleware/br"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
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
	ErrCtx    = errors.New("echo context is nil")
	ErrData   = errors.New("data interface is nil")
	ErrFS     = errors.New("embed filesystem instance is empty")
	ErrName   = errors.New("template name string is empty")
	ErrPorts  = errors.New("the server ports are not configured")
	ErrRoutes = errors.New("e echo instance is nil")
	ErrTmpl   = errors.New("named template cannot be found")
	ErrW      = errors.New("w io.writer instance is nil")
	ErrZap    = errors.New("zap logger instance is nil")
)

// Configuration of the handler.
type Configuration struct {
	Environment config.Config // Environment configurations from the host system.
	Brand       *[]byte       // Brand contains the Defacto2 ASCII logo.
	Public      embed.FS      // Public facing files.
	View        embed.FS      // View contains Go templates.
	Version     string        // Version is the results of GoReleaser build command.
	RecordCount int           // The total number of file records in the database.
}

// Controller is the primary instance of the Echo router.
func (c Configuration) Controller(logger *zap.SugaredLogger) *echo.Echo {
	configs := c.Environment

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = configs.CustomErrorHandler

	if tmpl, err := c.Registry(logger); err != nil {
		logger.Fatal(err)
	} else {
		e.Renderer = tmpl
	}

	middlewares := []echo.MiddlewareFunc{
		middleware.Rewrite(rewrites()),
		middleware.NonWWWRedirect(),
	}
	if httpsRedirect := configs.HTTPSRedirect && configs.TLSPort > 0; httpsRedirect {
		middlewares = append(middlewares, middleware.HTTPSRedirect())
	}
	e.Pre(middlewares...)

	// *************************************************
	//  NOTE: NEVER USE the middleware.Timeout()
	//   It is broken and should not be in the
	//   labstack/echo library, as it can easily crash!
	// *************************************************
	middlewares = []echo.MiddlewareFunc{
		middleware.Secure(),
		middleware.RequestLoggerWithConfig(c.configZapLogger()),
		c.NoCrawl,
		middleware.RemoveTrailingSlashWithConfig(configRTS()),
	}
	switch strings.ToLower(configs.Compression) {
	case "gzip":
		middlewares = append(middlewares, middleware.Gzip())
	case "br":
		middlewares = append(middlewares, br.Brotli())
	}
	if configs.ProductionMode {
		middlewares = append(middlewares, middleware.Recover()) // recover from panics
	}
	e.Use(middlewares...)

	e = EmbedDirs(e, c.Public)
	e = MovedPermanently(e)
	e = htmxGroup(e,
		logger,
		c.Environment.ProductionMode,
		c.Environment.DownloadDir)
	e, err := c.FilesRoutes(e, logger, c.Public)
	if err != nil {
		logger.Fatal(err)
	}
	group := html3.Routes(e, logger)
	group.GET(Downloader, func(cx echo.Context) error {
		return c.downloader(cx, logger)
	})
	return e
}

// EmbedDirs serves the static files from the directories embed to the binary.
func EmbedDirs(e *echo.Echo, currentFs fs.FS) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	dirs := map[string]string{
		"/image/artpack":   "public/image/artpack",
		"/image/html3":     "public/image/html3",
		"/image/layout":    "public/image/layout",
		"/image/milestone": "public/image/milestone",
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
func (c Configuration) Info(logger *zap.SugaredLogger) {
	w := bufio.NewWriter(os.Stdout)
	nr := bytes.NewReader(*c.Brand)
	if l, err := io.Copy(w, nr); err != nil {
		logger.Warnf("Could not print the brand logo: %s.", err)
	} else if l > 0 {
		fmt.Fprint(w, "\n\n")
		w.Flush()
	}

	fmt.Fprintf(w, "  %s.\n", cmd.Copyright())
	fmt.Fprintf(w, "%s\n", c.versionBrief())

	cpuInfo := fmt.Sprintf("  %d active routines sharing %d usable threads on %d CPU cores.",
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	fmt.Fprintln(w, cpuInfo)

	golangInfo := fmt.Sprintf("  compiled with Go %s for %s on %s.\n",
		runtime.Version()[2:], cmd.OS(), cmd.Arch())
	fmt.Fprintln(w, golangInfo)
	//
	// All additional feedback should go in internal/config/check.go (c *Config) Checks()
	//
	w.Flush()
}

// PortErr handles the error when the HTTP or HTTPS server cannot start.
func (c Configuration) PortErr(logger *zap.SugaredLogger, port uint, err error) {
	s := "HTTP"
	if port == c.Environment.TLSPort {
		s = "TLS"
	}
	var portErr *net.OpError
	switch {
	case !c.Environment.ProductionMode && errors.As(err, &portErr):
		logger.Infof("air or task server could not start (this can probably be ignored): %s.", err)
	case errors.Is(err, net.ErrClosed),
		errors.Is(err, http.ErrServerClosed):
		logger.Infof("%s server shutdown gracefully.", s)
	case errors.Is(err, os.ErrNotExist):
		logger.Fatalf("%s server on port %d could not start: %w.", s, port, err)
	default:
	}
}

// Registry returns the template renderer.
func (c Configuration) Registry(logger *zap.SugaredLogger) (*TemplateRegistry, error) {
	webapp := app.Templ{
		Environment: c.Environment,
		Brand:       *c.Brand,
		Public:      c.Public,
		Version:     c.Version,
		View:        c.View,
	}
	tmpls, err := webapp.Templates()
	if err != nil {
		return nil, fmt.Errorf("webapp.Templates: %w", err)
	}
	src := html3.Templates(logger, c.View)
	maps.Copy(tmpls, src)
	src = htmx.Templates(c.View)
	maps.Copy(tmpls, src)
	return &TemplateRegistry{Templates: tmpls}, nil
}

// ShutdownHTTP waits for a Ctrl-C keyboard press to initiate a graceful shutdown of the HTTP web server.
// The shutdown procedure occurs a few seconds after the key press.
func (c *Configuration) ShutdownHTTP(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(ErrRoutes)
	}
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	waitDuration := ShutdownWait
	waitCount := ShutdownCounter
	ticker := 1 * time.Second
	if c.Environment.LocalMode {
		waitDuration = 0
		waitCount = 0
		ticker = 1 * time.Millisecond // this cannot be zero
	}
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer func() {
		const alert = "Detected Ctrl + C, server will shutdown"
		_ = logger.Sync() // do not check Sync errors as there can be false positives
		dst := os.Stdout
		w := bufio.NewWriter(dst)
		fmt.Fprintf(w, "\n%s in %v ", alert, waitDuration)
		w.Flush()
		count := waitCount
		pause := time.NewTicker(ticker)
		for range pause.C {
			count--
			w := bufio.NewWriter(dst)
			if count <= 0 {
				fmt.Fprintf(w, "\r%s %s\n", alert, "now     ")
				w.Flush()
				break
			}
			fmt.Fprintf(w, "\r%s in %ds ", alert, count)
			w.Flush()
		}
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
		if err := e.Shutdown(ctx); err != nil {
			logger.Fatalf("Server shutdown caused an error: %w.", err)
		}
		logger.Infoln("Server shutdown complete.")
		_ = logger.Sync()
		signal.Stop(quit)
		cancel()
	}()
}

// Start the HTTP, and-or the TLS servers.
func (c *Configuration) Start(e *echo.Echo, logger *zap.SugaredLogger, configs config.Config) error {
	if e == nil {
		panic(ErrRoutes)
	}
	switch {
	case configs.UseTLS() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			c.StartHTTP(e2, logger)
		}()
		go c.StartTLS(e, logger)
	case configs.UseTLSLocal() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			c.StartHTTP(e2, logger)
		}()
		go c.StartTLSLocal(e, logger)
	case configs.UseTLS():
		go c.StartTLS(e, logger)
	case configs.UseHTTP():
		go c.StartHTTP(e, logger)
	case configs.UseTLSLocal():
		go c.StartTLSLocal(e, logger)
	default:
		return ErrPorts
	}
	return nil
}

// StartHTTP starts the insecure HTTP web server.
func (c *Configuration) StartHTTP(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(ErrRoutes)
	}
	port := c.Environment.HTTPPort
	if port == 0 {
		return
	}
	address := fmt.Sprintf(":%d", port)
	if err := e.Start(address); err != nil {
		c.PortErr(logger, port, err)
	}
}

// StartTLS starts the encrypted TLS web server.
func (c *Configuration) StartTLS(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(ErrRoutes)
	}
	port := c.Environment.TLSPort
	if port == 0 {
		return
	}
	cert := c.Environment.TLSCert
	key := c.Environment.TLSKey
	const failure = "Could not start the TLS server"
	if cert == "" || key == "" {
		logger.Fatalf("%s, missing certificate or key file.", failure)
	}
	if !helper.File(cert) {
		logger.Fatalf("%s, certificate file does not exist: %s.", failure, cert)
	}
	if !helper.File(key) {
		logger.Fatalf("%s, key file does not exist: %s.", failure, key)
	}
	address := fmt.Sprintf(":%d", port)
	if err := e.StartTLS(address, "", ""); err != nil {
		c.PortErr(logger, port, err)
	}
}

// StartTLSLocal starts the localhost, encrypted TLS web server.
// This should only be triggered when the server is running in local mode.
func (c *Configuration) StartTLSLocal(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(ErrRoutes)
	}
	port := c.Environment.TLSPort
	if port == 0 {
		return
	}
	const cert, key = "public/certs/cert.pem", "public/certs/key.pem"
	const failure = "Could not read the internal localhost"
	cpem, err := c.Public.ReadFile(cert)
	if err != nil {
		logger.Fatalf("%s, TLS certificate: %s.", failure, err)
	}
	kpem, err := c.Public.ReadFile(key)
	if err != nil {
		logger.Fatalf("%s, TLS key: %s.", failure, err)
	}
	lock := strings.TrimSpace(c.Environment.TLSHost)
	var address string
	const showAllConnections = ""
	switch lock {
	case showAllConnections:
		address = fmt.Sprintf(":%d", port)
	default:
		address = fmt.Sprintf("%s:%d", lock, port)
	}
	if err := e.StartTLS(address, cpem, kpem); err != nil {
		c.PortErr(logger, port, err)
	}
}

// downloader route for the file download handler under the html3 group.
func (c Configuration) downloader(cx echo.Context, logger *zap.SugaredLogger) error {
	d := download.Download{
		Inline: false,
		Path:   c.Environment.DownloadDir,
	}
	if err := d.HTTPSend(cx, logger); err != nil {
		return fmt.Errorf("d.HTTPSend: %w", err)
	}
	return nil
}

// versionBrief returns the application version string.
func (c Configuration) versionBrief() string {
	if c.Version == "" {
		return "  no version info, app compiled binary directly."
	}
	return fmt.Sprintf("  %s.", cmd.Commit(c.Version))
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
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	const layout = "layout"
	if name == "" {
		return ErrName
	}
	if w == nil {
		return fmt.Errorf("%w: %w", echo.ErrRendererNotRegistered, ErrW)
	}
	if data == nil {
		return fmt.Errorf("%w: %w", echo.ErrRendererNotRegistered, ErrData)
	}
	if c == nil {
		return fmt.Errorf("%w: %w", echo.ErrRendererNotRegistered, ErrCtx)
	}
	tmpl, exists := t.Templates[name]
	if !exists {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	if err := tmpl.ExecuteTemplate(w, layout, data); err != nil {
		return fmt.Errorf("tmpl.ExecuteTemplate layout: %w", err)
	}
	return nil
}
