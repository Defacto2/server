// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
package handler

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"runtime"
	"strconv"

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
	"github.com/Defacto2/server/internal/nils"
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
	ErrNoName = errors.New("handler: name is empty")
	ErrNoTmpl = errors.New("handler: named template cannot be found")
	ErrNoTLS  = errors.New("handler: tls usage requires certificate and key files")
	ErrNoPort = errors.New("handler: web server ports are not configured")
)

// Server of the handler.
type Server struct {
	Public      fs.FS            // Public facing files.
	View        fs.FS            // View contains Go templates.
	Version     string           // Version is the results of GoReleaser build command.
	Brand       []byte           // Brand contains the Defacto2 ASCII logo.
	Environment config.Config    // Environment configurations from the host system.
	RecordCount int64            // The total number of file records in the database.
	TidbitIndex fulltext.Tidbits // Fulltext search index of the tidbit markdown files.
}

// Handler is the primary instance of the Echo router.
func (serv *Server) Handler(ctx context.Context, sl *slog.Logger, db *sql.DB) *echo.Echo { //nolint:funlen
	const msg = "controller handler"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	logFatal := func(k, v string, err error) {
		logs.Fatal(ctx, sl, msg, slog.String(k, v), slog.Any("fatal", err))
	}
	logRoute := func(m, p string) {
		sl.Info("route", slog.String("method", m), slog.String("path", p))
	}

	settings := serv.Environment
	prodMode := bool(settings.ProdMode)

	httpErr := func(c *echo.Context, err error) {
		config.CustomErrorHandler(sl, c, err)
	}

	onAddRoute := func(route echo.Route) error {
		if !prodMode {
			return nil
		}
		logRoute(route.Method, route.Path)
		return nil
	}

	templates, err := serv.TemplRegistry(ctx, sl, db)
	if err != nil {
		logFatal("template", "cannot register templates", err)
	}

	const setAs16MB = 16 * 1024 * 1024
	config := echo.Config{
		Logger:           sl,
		HTTPErrorHandler: httpErr,
		Router: echo.NewRouter(echo.RouterConfig{
			AllowOverwritingRoute:     false,
			AutoHandleHEAD:            false,
			MethodNotAllowedHandler:   nil,
			NotFoundHandler:           nil,
			OptionsMethodHandler:      nil,
			UnescapePathParamValues:   false,
			UseEscapedPathForMatching: false,
		}),
		OnAddRoute:                      onAddRoute,
		Filesystem:                      nil,
		Binder:                          nil,
		Validator:                       nil,
		Renderer:                        templates,
		JSONSerializer:                  nil,
		IPExtractor:                     nil,
		FormParseMaxMemory:              setAs16MB,
		EnablePathUnescapingStaticFiles: false, // false is the new default after Echo v5.2.1
		// a breaking changed introduced in echo v5.3.0, using the default false breaks our app
		NoGroupAutoRegister404Routes: true,
	}

	e := echo.NewWithConfig(config)
	if settings.LogAll {
		// echo prefix options that get used by RequestLoggerConfig
		pprof.Register(e)
	}

	// pre middleware
	e.Pre(
		middleware.Rewrite(rewrites()),
		middleware.NonWWWRedirect(),
	)

	// use middleware
	if settings.Compression {
		e.Use(middleware.Gzip())
	}
	if settings.ProdMode {
		e.Use(middleware.Recover())
	}
	e.Use(
		middleware.Secure(),
		middleware.RequestLoggerWithConfig(serv.RequestLoggerConfig(sl)),
		serv.NoCrawl,
		middleware.RemoveTrailingSlashWithConfig(configTrailSlash()),
	)

	// browser paths and routes
	e = RouteFS(e, serv.Public)
	e = RouteMoved(e)
	configHTMX{
		prodMode: prodMode,
		download: dir.Directory(serv.Environment.AbsDownload),
	}.routeHTMX(sl, e, db)
	e, err = serv.RouteFS(sl, e, db, serv.Public)
	if err != nil {
		logFatal("file route", "cannot register routes", err)
	}

	group := html3.Route(sl, e, db)
	group.GET(Downloader, func(c *echo.Context) error {
		return serv.downloader(sl, c, db)
	})

	return e
}

// RouteFS serves the static files from the directories embed to the binary.
func RouteFS(e *echo.Echo, fsys fs.FS) *echo.Echo {
	const msg = "embed dirs handler"
	if err := nils.Check(e, fsys); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	missing := func(_ *echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "directory not found")
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
	// allows the files to be served, but return a 404 for root directories.
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(fsys, fsRoot))
		e.GET(path, missing)
	}

	return e
}

// Print the application logo and software information to the w Writer.
func (serv *Server) Print(w io.Writer) (errs error) {
	if w == nil {
		w = io.Discard
	}

	nr := bytes.NewReader(serv.Brand)
	n, err := io.Copy(w, nr)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	if n > 0 {
		_, err := fmt.Fprint(w, "\n\n")
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	_, err = fmt.Fprintf(w, "  %s.\n", flags.Copyright())
	if err != nil {
		errs = errors.Join(errs, err)
	}

	_, err = fmt.Fprintf(w, "%s\n", serv.versionBrief())
	if err != nil {
		errs = errors.Join(errs, err)
	}

	fs := "  %d active routines sharing %d usable threads on %d CPU cores."
	cpuInfo := fmt.Sprintf(fs, runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	_, err = fmt.Fprintln(w, cpuInfo)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	fs = "  Compiled with Go v%s for %s on %s."
	golangInfo := fmt.Sprintf(fs, runtime.Version()[2:], flags.OS(), flags.Arch())
	_, err = fmt.Fprintln(w, golangInfo)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	to, fr, _, s, err := helper.DiskStat("/")
	if err != nil {
		errs = errors.Join(errs, err)
	}

	total := int64(to)
	free := int64(fr)
	// Disk (/): 300.28 GiB / 464.17 GiB (65%)
	diskInfo := fmt.Sprintf("  Disk (/) %s / %s (%s).\n",
		helper.ByteCount(free), helper.ByteCount(total),
		s)
	_, err = fmt.Fprintln(w, diskInfo)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
	//
	// All additional feedback should go in internal/config/check.go (c *Config) Checks()
	//
}

// TemplRegistry returns the template registry for the renderer.
func (serv *Server) TemplRegistry(ctx context.Context, sl *slog.Logger, db *sql.DB) (*TemplateRegistry, error) {
	const format = "template registry handler: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}

	webapp := app.Templ{
		Public:      serv.Public,
		View:        serv.View,
		Subresource: app.SRI{}, //nolint:exhaustruct // SRI fields are computed via Verify() method
		Version:     serv.Version,
		Brand:       serv.Brand,
		Environment: serv.Environment,
		RecordCount: serv.RecordCount,
	}
	tmpls, err := webapp.Templates(ctx, db)
	if err != nil {
		return nil, fmt.Errorf(format, err)
	}

	// copy HTML3 templates
	src := html3.Templates(ctx, sl, db, serv.View)
	maps.Copy(tmpls, src)

	// copy HTMX templates
	src = htmx.Templates(serv.View)
	maps.Copy(tmpls, src)

	return &TemplateRegistry{Templates: tmpls}, nil
}

// EchoConfig returns the base server start configuration.
func (serv *Server) EchoConfig() echo.StartConfig {
	config := echo.StartConfig{ //nolint:exhaustruct
		HideBanner: true,
		HidePort:   true,
	}

	return config
}

// Start the HTTP, and-or the TLS servers that serves the web application.
func (serv *Server) Start(ctx context.Context, sl *slog.Logger, h http.Handler, configs config.Config) error {
	const format = "start server handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	var err error
	switch {
	case configs.UseTLS() && configs.UseHTTP():
		err = serv.StartDual(ctx, sl, h)
	case configs.UseLocal() && configs.UseHTTP():
		err = serv.StartLocals(ctx, sl, h)
	case configs.UseTLS():
		err = serv.StartTLS(ctx, sl, h)
	case configs.UseHTTP():
		err = serv.StartHTTP(ctx, sl, h)
	case configs.UseLocal():
		err = serv.StartLocal(ctx, sl, h)
	default:
		return fmt.Errorf(format, ErrNoPort)
	}
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

// StartLocals is intended for development only. It starts the HTTP server and plus the TLS server.
// However, the TLS uses an unsigned certificate and key file that is unusable on the Internet.
func (serv *Server) StartLocals(ctx context.Context, sl *slog.Logger, h http.Handler) error {
	const format = "start locals handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	return serv.startDual(ctx, sl, h, true)
}

// StartDual is intended for production only. It starts the HTTP server and plus the TLS server.
// However, the TLS must be used with a valid, signed certificate and key file that is suitable on the Internet.
func (serv *Server) StartDual(ctx context.Context, sl *slog.Logger, h http.Handler) error {
	const format = "start dual handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	return serv.startDual(ctx, sl, h, false)
}

// HTTP returns the unencrypted HTTP server configuration.
func (serv *Server) HTTP() echo.StartConfig {
	config := serv.EchoConfig()

	port := serv.Environment.HTTPPort.Value()
	address := serv.address(port)
	if address == "" {
		return config
	}

	config.Address = address

	return config
}

// Local is intended for development only.
// It returns the TLS server configuration and the content of an unsigned certificate
// and key file that is unusable on the Internet.
//
// Any returned errors should be fatal.
func (serv *Server) Local(ctx context.Context, sl *slog.Logger) (
	config echo.StartConfig, cert []byte, key []byte, err error,
) {
	const format = "local handler configuration: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return config, cert, key, fmt.Errorf(format, err)
	}

	config = serv.EchoConfig()
	port := serv.Environment.TLSPort.Value()
	address := serv.address(port)
	if address == "" {
		return config, cert, key, nil
	}
	config.Address = address

	const localCert = "public/certs/cert.pem"
	cert, err = fs.ReadFile(serv.Public, localCert)
	if err != nil {
		return config, cert, key, fmt.Errorf(format, err)
	}

	const localKey = "public/certs/key.pem"
	key, err = fs.ReadFile(serv.Public, localKey)
	if err != nil {
		return config, cert, key, fmt.Errorf(format, err)
	}

	return config, cert, key, nil
}

// TLS returns server configuration and the content of the certificate and key file.
//
// Any returned errors should be fatal.
func (serv *Server) TLS(ctx context.Context, sl *slog.Logger) (
	config echo.StartConfig, cert []byte, key []byte, err error,
) {
	const format = "tls handler configuration: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return config, cert, key, fmt.Errorf(format, err)
	}

	config = serv.EchoConfig()
	port := serv.Environment.TLSPort.Value()
	address := serv.address(port)
	if address == "" {
		return config, cert, key, nil
	}
	config.Address = address

	certPath := serv.Environment.TLSCert
	keyPath := serv.Environment.TLSKey
	if certPath == "" || keyPath == "" {
		return config, cert, key, fmt.Errorf(format, ErrNoTLS)
	}
	if !helper.File(certPath.String()) {
		return config, cert, key, fmt.Errorf(format+" certificate", ErrNoTLS)
	}
	if !helper.File(keyPath.String()) {
		return config, cert, key, fmt.Errorf(format+" key file", ErrNoTLS)
	}

	cert, err = os.ReadFile(certPath.String())
	if err != nil {
		return config, cert, key, fmt.Errorf(format, err)
	}
	key, err = os.ReadFile(keyPath.String())
	if err != nil {
		return config, cert, key, fmt.Errorf(format, err)
	}

	return config, cert, key, nil
}

// StartHTTP starts the unencrypted HTTP web server.
//
// The default port for the HTTP protocol is 80.
func (serv *Server) StartHTTP(ctx context.Context, sl *slog.Logger, h http.Handler) error {
	const format = "start http handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	httpConfig := serv.HTTP()
	sl.Info("Starting HTTP Listener",
		slog.String("address", httpConfig.Address))

	err := httpConfig.Start(ctx, h)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTP Server crashed unexpectedly",
			slog.Any("error", err))
		return fmt.Errorf(format, err)
	}

	return nil
}

// StartTLS starts the encrypted TLS web server.
//
// The default port for the HTTPS protocol is 443.
func (serv *Server) StartTLS(ctx context.Context, sl *slog.Logger, h http.Handler) error {
	const format = "start tls handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	httpsConfig, certFile, keyFile, err := serv.TLS(ctx, sl)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	sl.Info("Starting HTTPS Listener", slog.String("address", httpsConfig.Address))

	// Point to the valid SSL or TLS files
	err = httpsConfig.StartTLS(ctx, h, certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTPS Server crashed unexpectedly",
			slog.Any("error", err))
		return fmt.Errorf(format, err)
	}

	return nil
}

// StartLocal is intended for development only. It starts the TLS server.
// However, the TLS configuration uses an unsigned certificate and key file that is unusable on the Internet.
func (serv *Server) StartLocal(ctx context.Context, sl *slog.Logger, h http.Handler) error {
	const format = "start local tls handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	httpsConfig, certFile, keyFile, err := serv.Local(ctx, sl)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	sl.Info("Starting HTTPS Listener",
		slog.String("address", httpsConfig.Address))

	err = httpsConfig.StartTLS(ctx, h, certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sl.Error("HTTPS Server crashed unexpectedly",
			slog.Any("error", err))
		return fmt.Errorf(format, err)
	}

	return nil
}

func (serv *Server) startDual(ctx context.Context, sl *slog.Logger, h http.Handler, local bool) error {
	const format = "start dual handler: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	g, ctx := errgroup.WithContext(ctx)

	httpConfig := serv.HTTP()
	tlsConfig := echo.StartConfig{} //nolint:exhaustruct
	certB, keyB := []byte{}, []byte{}
	var err error
	if local {
		tlsConfig, certB, keyB, err = serv.Local(ctx, sl)
	} else {
		tlsConfig, certB, keyB, err = serv.TLS(ctx, sl)
	}
	if err != nil {
		return fmt.Errorf(format, err)
	}

	g.Go(func() error {
		sl.Info("Starting HTTP Listener", slog.String("address", httpConfig.Address))

		err := httpConfig.Start(ctx, h)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sl.Error("HTTP Server crashed unexpectedly",
				slog.Any("error", err))
			return fmt.Errorf(format, err)
		}

		return nil
	})

	g.Go(func() error {
		sl.Info("Starting HTTPS Listener", slog.String("address", tlsConfig.Address))

		err := tlsConfig.StartTLS(ctx, h, certB, keyB)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sl.Error("HTTPS Server crashed unexpectedly",
				slog.Any("error", err))
			return fmt.Errorf(format, err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		sl.Error("System tracking intercepted a service failure",
			slog.Any("error", err))
		return fmt.Errorf(format, err)
	}

	sl.Info("Dual server infrastructure successfully stopped.")
	return nil
}

func (serv *Server) address(port uint16) string {
	if port == 0 {
		return ""
	}

	p := strconv.Itoa(int(port))
	if host := serv.Environment.MatchHost; host != "" {
		return string(host) + ":" + p
	}

	return ":" + p
}

// downloader is used by the html3 group route as the file download handler.
func (serv *Server) downloader(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "downloader htm3 group handler: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	d := download.Download{
		Inline: false,
		Dir:    dir.Directory(serv.Environment.AbsDownload),
	}
	if err := d.HTTPSend(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

// versionBrief returns the application version string.
func (serv *Server) versionBrief() string {
	ver := serv.Version
	if ver == "" {
		return "  no version info, app compiled binary directly."
	}

	return "  " + flags.VersionCommit(ver) + "."
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
	const format = "template registry render handler %s: %w"
	if err := nils.Check(c, w, data); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	if name == "" {
		return fmt.Errorf(format, "name layout", ErrNoName)
	}

	tmpl, exists := t.Templates[name]
	if !exists {
		return fmt.Errorf(format, name, ErrNoTmpl)
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		return fmt.Errorf(format, "execute template: "+name, err)
	}

	return nil
}
