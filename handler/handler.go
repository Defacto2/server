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
	"net/http"
	"runtime"

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
func (c *Configuration) Handler(ctx context.Context, sl *slog.Logger, db *sql.DB) *echo.Echo { //nolint:funlen
	const msg = "controller handler"
	err := panics.SD(sl, db)
	if err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	envConfig := c.Environment
	prodMode := bool(envConfig.ProdMode)

	httpErr := func(ec *echo.Context, err error) {
		config.CustomErrorHandler(ctx, sl, ec, err)
	}

	onAddRoute := func(route echo.Route) error {
		if !prodMode {
			return nil
		}
		sl.Info(
			"route",
			slog.String("method", route.Method),
			slog.String("path", route.Path),
		)
		return nil
	}

	templates, err := c.TemplRegistry(ctx, sl, db)
	if err != nil {
		logs.Fatal(ctx, sl, msg,
			slog.String("template", "could not register the templates"),
			slog.Any("fatal", err))
	}

	const setAs16MB = 16 * 1024 * 1024
	echoConfig := echo.Config{
		Logger:           sl,
		HTTPErrorHandler: httpErr,
		Router: echo.NewRouter(echo.RouterConfig{
			NotFoundHandler:           nil,
			MethodNotAllowedHandler:   nil,
			OptionsMethodHandler:      nil,
			AllowOverwritingRoute:     false,
			UnescapePathParamValues:   false,
			UseEscapedPathForMatching: false,
		}),
		OnAddRoute:         onAddRoute,
		Filesystem:         nil,
		Binder:             nil,
		Validator:          nil,
		Renderer:           templates,
		JSONSerializer:     nil,
		IPExtractor:        nil,
		FormParseMaxMemory: setAs16MB,
	}

	e := echo.NewWithConfig(echoConfig)
	if envConfig.LogAll {
		// echo prefix options that get used by RequestLoggerConfig
		pprof.Register(e)
	}
	// pre middleware
	e.Pre(
		middleware.Rewrite(rewrites()),
		middleware.NonWWWRedirect(),
	)
	// use middleware
	if envConfig.Compression {
		e.Use(middleware.Gzip())
	}
	if envConfig.ProdMode {
		e.Use(middleware.Recover())
	}
	e.Use(
		middleware.Secure(),
		middleware.RequestLoggerWithConfig(c.RequestLoggerConfig(sl)),
		c.NoCrawl,
		middleware.RemoveTrailingSlashWithConfig(configTrailSlash()),
	)
	// browser paths and routes
	e = AppendEmbed(e, c.Public)
	e = AppendMoved(e)
	ch := configHtmx{
		prodMode: prodMode,
		download: dir.Directory(c.Environment.AbsDownload),
	}
	e = ch.append(ctx, sl, e, db)
	e, err = c.AppendFiles(ctx, sl, e, db, c.Public)
	if err != nil {
		logs.Fatal(ctx, sl, msg,
			slog.String("file routes", "could not register the routes"),
			slog.Any("fatal", err))
	}
	group := html3.Routes(sl, e, db)
	group.GET(Downloader, func(ec *echo.Context) error {
		return c.downloader(ctx, sl, ec, db)
	})
	return e
}

// AppendEmbed serves the static files from the directories embed to the binary.
func AppendEmbed(e *echo.Echo, currentFs fs.FS) *echo.Echo {
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
	notfound := func(_ *echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "directory not found")
	}
	// Allows files to be served but returns 404 for the root directories.
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(currentFs, fsRoot))
		e.GET(path, notfound)
	}
	return e
}

// Print the application logo and information to the w Writer.
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
	format := "  %d active routines sharing %d usable threads on %d CPU cores."
	cpuInfo := fmt.Sprintf(format, runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	_, _ = fmt.Fprintln(w, cpuInfo)
	format = "  Compiled on Go %s for %s with %s."
	golangInfo := fmt.Sprintf(format, runtime.Version()[2:], flags.OS(), flags.Arch())
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

// TemplRegistry returns the template registry for the renderer.
func (c *Configuration) TemplRegistry(ctx context.Context, sl *slog.Logger, db *sql.DB) (*TemplateRegistry, error) {
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
	tmpls, err := webapp.Templates(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	src := html3.Templates(ctx, db, sl, c.View)
	maps.Copy(tmpls, src)
	src = htmx.Templates(c.View)
	maps.Copy(tmpls, src)
	return &TemplateRegistry{Templates: tmpls}, nil
}

func (c *Configuration) EchoConfig() echo.StartConfig {
	config := echo.StartConfig{ //nolint:exhaustruct
		HideBanner: true,
		HidePort:   true,
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

func (c *Configuration) Local(ctx context.Context, sl *slog.Logger) (echo.StartConfig, string, string) {
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
		logs.Fatal(ctx, sl, msg,
			slog.String("certificate", "read file failure"), slog.Any("error", err))
	}
	keyFile, err := c.Public.ReadFile("public/certs/key.pem")
	if err != nil {
		logs.Fatal(ctx, sl, msg,
			slog.String("key", "read file failure"), slog.Any("error", err))
	}
	return config, string(certFile), string(keyFile)
}

func (c *Configuration) TLS(ctx context.Context, sl *slog.Logger) (echo.StartConfig, string, string) {
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
		logs.Fatal(ctx, sl, msg,
			slog.String("failure", "missing critical file"),
			slog.String("certificate file", string(certFile)),
			slog.String("key file", string(keyFile)))
	}
	if !helper.File(certFile.String()) {
		logs.Fatal(ctx, sl, msg,
			slog.String("certificate file", "file does not exist"))
	}
	if !helper.File(keyFile.String()) {
		logs.Fatal(ctx, sl, msg,
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
	sl.Info("Starting HTTP Listener",
		slog.String("address", httpConfig.Address))
	err := httpConfig.Start(ctx, h)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTP Server crashed unexpectedly",
			slog.Any("error", err))
		return
	}
}

// StartTLS starts the encrypted TLS web server.
func (c *Configuration) StartTLS(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start tls handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	httpsConfig, certFile, keyFile := c.TLS(ctx, sl)
	sl.Info("Starting HTTPS Listener",
		slog.String("address", httpsConfig.Address))
	// Point to your valid SSL/TLS files
	err := httpsConfig.StartTLS(ctx, h, certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTPS Server crashed unexpectedly",
			slog.Any("error", err))
		return
	}
}

// StartLocal starts the insecure localhost, encrypted TLS web server.
// This should only be triggered when the server is running in local mode.
func (c *Configuration) StartLocal(ctx context.Context, sl *slog.Logger, h http.Handler) {
	const msg = "start local tls handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}

	httpsConfig, certFile, keyFile := c.Local(ctx, sl)
	sl.Info("Starting HTTPS Listener",
		slog.String("address", httpsConfig.Address))
	// Point to your valid SSL/TLS files
	err := httpsConfig.StartTLS(ctx, h, certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTPS Server crashed unexpectedly",
			slog.Any("error", err))
		return
	}
}

func (c *Configuration) startDual(ctx context.Context, sl *slog.Logger, h http.Handler, local bool) {
	const msg = "start dual handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	g, ctx := errgroup.WithContext(ctx)

	httpConfig := c.HTTP()
	httpsConfig := echo.StartConfig{} //nolint:exhaustruct
	certFile, keyFile := "", ""
	if local {
		httpsConfig, certFile, keyFile = c.Local(ctx, sl)
	} else {
		httpsConfig, certFile, keyFile = c.TLS(ctx, sl)
	}

	g.Go(func() error {
		sl.Info("Starting HTTP Listener",
			slog.String("address", httpConfig.Address))
		err := httpConfig.Start(ctx, h)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sl.Error("HTTP Server crashed unexpectedly",
				slog.Any("error", err))
			return fmt.Errorf("dual http server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		sl.Info("Starting HTTPS Listener",
			slog.String("address", httpsConfig.Address))
		// Point to your valid SSL/TLS files
		err := httpsConfig.StartTLS(ctx, h, certFile, keyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sl.Error("HTTPS Server crashed unexpectedly",
				slog.Any("error", err))
			return fmt.Errorf("dual https server: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		sl.Error("System tracking intercepted a service failure",
			slog.Any("error", err))
		return
	}

	sl.Info("Dual server infrastructure successfully stopped.")
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
func (c *Configuration) downloader(ctx context.Context, sl *slog.Logger, ec *echo.Context, db *sql.DB) error {
	const msg = "downloader htm3 group handler"
	if err := panics.SCD(sl, ec, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	d := download.Download{
		Inline: false,
		Dir:    dir.Directory(c.Environment.AbsDownload),
	}
	if err := d.HTTPSend(ctx, sl, ec, db); err != nil {
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
