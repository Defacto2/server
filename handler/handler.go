// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
package handler

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
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
	ErrName   = errors.New("template name string is empty")
	ErrRoutes = errors.New("e echo instance is nil")
	ErrTmpl   = errors.New("named template cannot be found")
	ErrW      = errors.New("w io.writer instance is nil")
	ErrZap    = errors.New("zap logger instance is nil")
)

// Configuration of the handler.
type Configuration struct {
	Import      *config.Config     // Import configurations from the host system environment.
	Logger      *zap.SugaredLogger // Logger is the zap sugared logger.
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	Public      embed.FS           // Public facing files.
	View        embed.FS           // View contains Go templates.
	Version     string             // Version is the results of GoReleaser build command.
	RecordCount int                // The total number of file records in the database.
}

// Controller is the primary instance of the Echo router.
func (c Configuration) Controller() *echo.Echo {
	logr := c.Logger
	// configurations
	e := echo.New()
	e.HideBanner = true                              // hide the Echo banner
	e.HTTPErrorHandler = c.Import.CustomErrorHandler // custom error handler (see: internal/config/logger.go)
	// register html templates
	templates, err := c.Registry()
	if err != nil {
		logr.Fatal(err)
	}
	e.Renderer = templates
	// middleware pre-configurations
	e.Pre(
		middleware.Rewrite(rewrites()), // rewrites for assets
		middleware.NonWWWRedirect(),    // redirect www.defacto2.net requests to defacto2.net
	)
	httpsRedirect := c.Import.HTTPSRedirect && c.Import.TLSPort > 0
	if httpsRedirect {
		e.Pre(middleware.HTTPSRedirect()) // http://defacto2.net to https://defacto2.net redirect
	}
	// middleware configurations
	// Note: NEVER USE the middleware.Timeout(),
	// it shouldn't be in the echo library and will cause server crashes
	e.Use(
		middleware.Secure(),       // XSS cross-site scripting protection
		c.Import.LoggerMiddleware, // custom HTTP logging middleware
		c.NoCrawl,                 // add X-Robots-Tag to all responses
		middleware.RemoveTrailingSlashWithConfig(configRTS()), // remove trailing slashes
	)
	switch strings.ToLower(c.Import.Compression) {
	case "gzip": // Gzip HTTP compression
		e.Use(middleware.Gzip())
	case "br": // experimental Brotli HTTP compression
		e.Use(br.Brotli())
	}
	if c.Import.ProductionMode {
		e.Use(middleware.Recover()) // recover from panics
	}
	// Static embedded web assets that get distributed in the binary
	e = c.EmbedDirs(e)
	// Routes for the web application
	// that first handles the permanent redirects
	e, err = c.Moved(e)
	if err != nil {
		logr.Fatal(err)
	}
	e, err = c.Routes(e, c.Public)
	if err != nil {
		logr.Fatal(err)
	}
	// Routes for the retro web tables
	retro := html3.Routes(logr, e)
	retro.GET(Downloader, c.downloader)
	// Routes for the htmx responses
	e = htmx.Routes(logr, e)
	return e
}

// EmbedDirs serves the static files from the directories embed to the binary.
func (c Configuration) EmbedDirs(e *echo.Echo) *echo.Echo {
	if e == nil {
		c.Logger.Fatal(ErrRoutes)
	}
	dirs := map[string]string{
		"/image/artpack":   "public/image/artpack",
		"/image/html3":     "public/image/html3",
		"/image/layout":    "public/image/layout",
		"/image/milestone": "public/image/milestone",
	}
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(c.Public, fsRoot))
		e.GET(path, func(_ echo.Context) error {
			return echo.NewHTTPError(http.StatusNotFound)
		})
	}
	return e
}

// Info prints the application information to the console.
func (c Configuration) Info() {
	w := bufio.NewWriter(os.Stdout)
	// Startup logo
	if logo := string(*c.Brand); len(logo) > 0 {
		w := bufio.NewWriter(os.Stdout)
		if _, err := fmt.Fprintf(w, "%s\n\n", logo); err != nil {
			c.Logger.Warnf("Could not print the brand logo: %s.", err)
		}
		w.Flush()
	}
	// Legal info
	fmt.Fprintf(w, "  %s.\n", cmd.Copyright())
	// Brief version
	fmt.Fprintf(w, "%s\n", c.version())
	// CPU info
	fmt.Fprintf(w, "  %d active routines sharing %d usable threads on %d CPU cores.\n",
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	// Go info
	fmt.Fprintf(w, "  compiled with Go %s for %s on %s.\n\n",
		runtime.Version()[2:], cmd.OS(), cmd.Arch())
	//
	// All additional feedback should go in internal/config/check.go (c *Config) Checks()
	//
	w.Flush()
}

// PortErr handles the error when the HTTP or HTTPS server cannot start.
func (c Configuration) PortErr(port uint, err error) {
	s := "HTTP"
	if port == c.Import.TLSPort {
		s = "TLS"
	}
	var portErr *net.OpError
	switch {
	case !c.Import.ProductionMode && errors.As(err, &portErr):
		c.Logger.Infof("air or task server could not start (this can probably be ignored): %s.", err)
	case errors.Is(err, net.ErrClosed),
		errors.Is(err, http.ErrServerClosed):
		c.Logger.Infof("%s server shutdown gracefully.", s)
	case errors.Is(err, os.ErrNotExist):
		c.Logger.Fatalf("%s server on port %d could not start: %w.", s, port, err)
	default:
	}
}

// Registry returns the template renderer.
func (c Configuration) Registry() (*TemplateRegistry, error) {
	webapp := app.Web{
		Import:  c.Import,
		Logger:  c.Logger,
		Brand:   c.Brand,
		Public:  c.Public,
		Version: c.Version,
		View:    c.View,
	}
	tmpls, err := webapp.Templates()
	if err != nil {
		return nil, err
	}
	src := html3.Templates(c.Logger, c.View)
	maps.Copy(tmpls, src)
	src = htmx.Templates(c.Logger, c.View)
	maps.Copy(tmpls, src)
	return &TemplateRegistry{Templates: tmpls}, nil
}

// ShutdownHTTP waits for a Ctrl-C keyboard press to initiate a graceful shutdown of the HTTP web server.
// The shutdown procedure occurs a few seconds after the key press.
func (c *Configuration) ShutdownHTTP(e *echo.Echo) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	waitDuration := ShutdownWait
	waitCount := ShutdownCounter
	ticker := 1 * time.Second
	if c.Import.LocalMode {
		waitDuration = 0
		waitCount = 0
		ticker = 1 * time.Millisecond // this cannot be zero
	}
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer func() {
		const alert = "Detected Ctrl-C, server will shutdown"
		_ = c.Logger.Sync() // do not check error as there's false positives
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
			c.Logger.Fatalf("Server shutdown caused an error: %w.", err)
		}
		c.Logger.Infoln("Server shutdown complete.")
		_ = c.Logger.Sync()
		signal.Stop(quit)
		cancel()
	}()
}

// StartHTTP starts the insecure HTTP web server.
func (c *Configuration) StartHTTP(e *echo.Echo) {
	port := c.Import.HTTPPort
	if port == 0 {
		return
	}
	address := fmt.Sprintf(":%d", port)
	if err := e.Start(address); err != nil {
		c.PortErr(port, err)
	}
}

// StartTLS starts the encrypted TLS web server.
func (c *Configuration) StartTLS(e *echo.Echo) {
	port := c.Import.TLSPort
	if port == 0 {
		return
	}
	cert := c.Import.TLSCert
	key := c.Import.TLSKey
	if cert == "" || key == "" {
		c.Logger.Fatalf("Could not start the TLS server, missing certificate or key file.")
	}
	if !helper.IsFile(cert) {
		c.Logger.Fatalf("Could not start the TLS server, certificate file does not exist: %s.", cert)
	}
	if !helper.IsFile(key) {
		c.Logger.Fatalf("Could not start the TLS server, key file does not exist: %s.", key)
	}
	address := fmt.Sprintf(":%d", port)
	if err := e.StartTLS(address, "", ""); err != nil {
		c.PortErr(port, err)
	}
}

// StartTLSLocal starts the localhost, encrypted TLS web server.
// This should only be triggered when the server is running in local mode.
func (c *Configuration) StartTLSLocal(e *echo.Echo) {
	port := c.Import.TLSPort
	if port == 0 {
		return
	}
	const cert, key = "public/certs/cert.pem", "public/certs/key.pem"
	cpem, err := c.Public.ReadFile(cert)
	if err != nil {
		c.Logger.Fatalf("Could not read the internal localhost, TLS certificate: %s.", err)
	}
	kpem, err := c.Public.ReadFile(key)
	if err != nil {
		c.Logger.Fatalf("Could not read the internal localhost, TLS key: %s.", err)
	}
	lock := strings.TrimSpace(c.Import.TLSHost)
	var address string
	switch lock {
	case "": // allow all connections
		address = fmt.Sprintf(":%d", port)
	default: // only allow connections from the
		address = fmt.Sprintf("%s:%d", lock, port)
	}
	if err := e.StartTLS(address, cpem, kpem); err != nil {
		c.PortErr(port, err)
	}
}

// downloader route for the file download handler under the html3 group.
func (c Configuration) downloader(ctx echo.Context) error {
	d := download.Download{
		Inline: false,
		Path:   c.Import.DownloadDir,
	}
	return d.HTTPSend(c.Logger, ctx)
}

// version returns the application version string.
func (c Configuration) version() string {
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
	tmpl, ok := t.Templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	return tmpl.ExecuteTemplate(w, layout, data)
}
