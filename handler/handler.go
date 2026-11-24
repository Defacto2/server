// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
//
//nolint:exhaustruct_v5,nonamedreturns
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
	"net"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"

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
	DBVersion   string           // DBVersion is the connected PostgreSQL server version.
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

	settings := serv.Environment
	prodMode := bool(settings.ProdMode)

	httpErr := func(c *echo.Context, err error) {
		config.CustomErrorHandler(sl, c, err)
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
		OnAddRoute:                      nil,
		Filesystem:                      os.DirFS("/"), // TODO: keep v4 behavour for now
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
	if !prodMode && settings.LogAll.Bool() {
		// Only expose /debug/pprof in development mode when verbose logging is enabled.
		pprof.Register(e)
	}

	// pre middleware
	e.Pre(
		middleware.Rewrite(rewrites()),
		middleware.NonWWWRedirect(),
	)

	// use middleware
	if settings.Compression {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: func(c *echo.Context) bool {
				return c.Path() == "/editor/configurations/pings"
			},
		}))
	}
	if settings.ProdMode {
		e.Use(middleware.Recover())
	}
	if settings.LogAll.Bool() {
		e.Use(middleware.RequestLoggerWithConfig(serv.RequestLoggerConfig(sl)))
	}
	e.Use(
		middleware.Secure(),
		serv.NoCrawl,
		middleware.RemoveTrailingSlashWithConfig(configTrailSlash()),
	)

	// browser paths and routes
	if err := serv.registerRoutes(sl, e, db); err != nil {
		logFatal("file route", "cannot register routes", err)
	}

	return e
}

func (serv *Server) registerRoutes(sl *slog.Logger, e *echo.Echo, db *sql.DB) error {
	e = RouteFS(e, serv.Public)
	e = RouteMoved(e)
	configHTMX{
		prodMode: bool(serv.Environment.ProdMode),
		download: dir.Directory(serv.Environment.AbsDownload),
		zooAuth:  serv.Environment.DemozooAuth.String(),
	}.routeHTMX(sl, e, db)
	e, err := serv.RouteFS(sl, e, db, serv.Public)
	if err != nil {
		return err
	}

	group := html3.Route(sl, e, db)
	group.GET(Downloader, func(c *echo.Context) error {
		return serv.downloader(sl, c, db)
	})

	return nil
}

// PrintRoutes registers the server's Echo routes without starting the HTTP listener
// and prints a sorted, aligned table of HTTP methods and paths to w.
func (serv *Server) PrintRoutes(w io.Writer, methodFilter string) error {
	if w == nil {
		w = os.Stdout
	}

	db, err := sql.Open("pgx", "")
	if err != nil {
		return fmt.Errorf("print routes placeholder db: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	e := echo.NewWithConfig(echo.Config{
		Logger:                       logs.Discard(),
		Filesystem:                   os.DirFS("/"),
		NoGroupAutoRegister404Routes: true,
	})
	if !serv.Environment.ProdMode.Bool() && serv.Environment.LogAll.Bool() {
		pprof.Register(e)
	}
	if err := serv.registerRoutes(logs.Discard(), e, db); err != nil {
		return fmt.Errorf("print routes: %w", err)
	}

	filter := strings.ToUpper(strings.TrimSpace(methodFilter))
	routes := slices.Clone(e.Router().Routes())
	if filter != "" {
		routes = slices.DeleteFunc(routes, func(r echo.RouteInfo) bool {
			return !strings.EqualFold(r.Method, filter)
		})
	}
	slices.SortFunc(routes, func(a, b echo.RouteInfo) int {
		if c := strings.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return strings.Compare(a.Method, b.Method)
	})

	color := logs.Color(w)
	style := func(code, s string) string {
		if !color || s == "" {
			return s
		}
		return code + s + "\033[0m"
	}
	methodStyle := func(m string) string {
		padded := fmt.Sprintf("%-7s", m)
		switch m {
		case http.MethodGet:
			return style("\033[32m", padded)
		case http.MethodPost:
			return style("\033[36m", padded)
		case http.MethodPut, http.MethodPatch:
			return style("\033[33m", padded)
		case http.MethodDelete:
			return style("\033[1;31m", padded)
		default:
			return style("\033[2m", padded)
		}
	}

	title := fmt.Sprintf("Web Server Routes (%d)", len(routes))
	if filter != "" {
		title = fmt.Sprintf("Web Server Routes (%d %s)", len(routes), filter)
	}
	_, _ = fmt.Fprintln(w, style("\033[1m", title))
	_, _ = fmt.Fprintln(w)
	for _, r := range routes {
		_, _ = fmt.Fprintf(w, "  %s %s\n", methodStyle(r.Method), r.Path)
	}

	return nil
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
	// Register GET before StaticFS so Echo v5.4.7+ does not auto-register the
	// directory prefix as a redirect route and collide with the 404 handler.
	for path, fsRoot := range dirs {
		e.GET(path, missing)
		e.StaticFS(path, echo.MustSubFS(fsys, fsRoot))
	}

	return e
}

// Print writes the Defacto2 startup banner, system info, status summary, and any config issues to w.
func (serv *Server) Print(w io.Writer) (errs error) { //nolint:funlen
	if w == nil {
		w = io.Discard
	}

	color := logs.Color(w)
	style := func(code, s string) string {
		if !color || s == "" {
			return s
		}
		return code + s + "\033[0m"
	}
	const (
		bold    = "\033[1m"
		dim     = "\033[2m"
		green   = "\033[32m"
		yellow  = "\033[33m"
		cyan    = "\033[36m"
		boldRed = "\033[1;31m"
	)
	row := func(label, val string) {
		_, err := fmt.Fprintf(w, "  %s %s\n", style(cyan, fmt.Sprintf("%-11s", label)), val)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if len(serv.Brand) > 0 {
		if _, err := io.Copy(w, bytes.NewReader(serv.Brand)); err != nil {
			errs = errors.Join(errs, err)
		}
		_, _ = fmt.Fprint(w, "\n\n")
	}
	_, _ = fmt.Fprintf(w, "  %s\n", style(dim, flags.Copyright()))

	// SYSTEM section
	_, _ = fmt.Fprintf(w, "\n%s\n", style(bold, "SYSTEM"))
	ver := serv.Version
	if ver == "" {
		ver = "dev"
	}
	row("Version", flags.Program+" "+ver)
	row("Runtime", fmt.Sprintf("Go %s (%s on %s, %d/%d CPU threads)",
		strings.TrimPrefix(runtime.Version(), "go"),
		flags.OS(), flags.Arch(), runtime.GOMAXPROCS(-1), runtime.NumCPU()))

	if to, fr, _, pct, err := helper.DiskStat("/"); err != nil {
		errs = errors.Join(errs, err)
	} else {
		row("Disk (/)", fmt.Sprintf("%s free / %s (%s)",
			helper.ByteCount(int64(fr)), helper.ByteCount(int64(to)), pct))
	}

	if groups, usr, err := helper.Owner(); err == nil && usr != "" {
		clean := slices.DeleteFunc(groups, func(e string) bool { return e == "" })
		if len(clean) > 0 {
			row("Process", fmt.Sprintf("%s (%s)", usr, strings.Join(clean, ", ")))
		} else {
			row("Process", usr)
		}
	}

	if pattern, usage, err := config.TempUsage(); err == nil {
		row("Temp dir", fmt.Sprintf("%s (%s)", pattern, usage))
	}

	// STATUS section
	_, _ = fmt.Fprintf(w, "\n%s\n", style(bold, "STATUS"))
	row("Database", serv.formatDBStatus(style, green, yellow, boldRed))
	row("Mode", serv.formatModeSummary())
	row("Listeners", serv.formatListeners())

	if addrs, err := serv.Environment.AllAddresses(); err == nil && len(addrs) > 0 {
		row("Addresses", style(cyan, addrs[0]))
		indent := strings.Repeat(" ", 14)
		for _, a := range addrs[1:] {
			_, _ = fmt.Fprintf(w, "%s%s\n", indent, style(cyan, a))
		}
	}

	// CONFIG ISSUES section (only when problems exist)
	if issues := serv.Environment.Issues(); len(issues) > 0 {
		_, _ = fmt.Fprintf(w, "\n%s\n", style(boldRed, "CONFIG ISSUES"))
		for _, iss := range issues {
			_, _ = fmt.Fprintf(w, "  %s %s — %s\n",
				style(boldRed, fmt.Sprintf("%-19s", iss.Env)),
				iss.Value, iss.Issue)
		}
	}

	_, _ = fmt.Fprintln(w)
	return errs
}

func (serv *Server) formatDBStatus(style func(string, string) string, green, yellow, boldRed string) string {
	if serv.DBVersion == "" && serv.RecordCount == 0 {
		return style(boldRed, "offline — limited functionality")
	}
	countStr := formatWithCommas(serv.RecordCount) + " artifacts"
	label := countStr
	if serv.DBVersion != "" {
		label = serv.DBVersion + " · " + countStr
	}
	if serv.RecordCount < config.MinimumFiles {
		return style(yellow, label+" (below minimum)")
	}
	return style(green, label)
}

func (serv *Server) formatModeSummary() string {
	env := serv.Environment
	parts := make([]string, 0, 5)
	if env.ProdMode.Bool() {
		parts = append(parts, "production")
	} else {
		parts = append(parts, "development")
	}
	if env.ReadOnly.Bool() {
		parts = append(parts, "read-only")
	} else {
		parts = append(parts, "write mode")
	}
	if env.NoCrawl.Bool() {
		parts = append(parts, "no-crawl")
	}
	if !env.Compression.Bool() {
		parts = append(parts, "gzip off")
	}
	if env.LogAll.Bool() {
		parts = append(parts, "log-all")
	}
	return strings.Join(parts, " · ")
}

func (serv *Server) formatListeners() string {
	env := serv.Environment
	parts := make([]string, 0, 2)
	if env.UseHTTP() {
		parts = append(parts, fmt.Sprintf("HTTP :%d", env.HTTPPort))
	}
	switch {
	case env.UseTLS():
		parts = append(parts, fmt.Sprintf("HTTPS :%d", env.TLSPort))
	case env.UseLocal():
		parts = append(parts, fmt.Sprintf("HTTPS :%d (self-signed)", env.TLSPort))
	}
	if len(parts) == 0 {
		return "none"
	}
	if host := env.MatchHost.String(); host != "" {
		return strings.Join(parts, " · ") + " (host " + host + ")"
	}
	return strings.Join(parts, " · ")
}

func formatWithCommas(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	rem := len(s) % 3
	if rem > 0 {
		b.WriteString(s[:rem])
	}
	for i := rem; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// TemplRegistry returns the template registry for the renderer.
func (serv *Server) TemplRegistry(ctx context.Context, sl *slog.Logger, db *sql.DB) (*TemplateRegistry, error) {
	const format = "template registry handler: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}

	webapp := app.WebApp{
		Public:      serv.Public,
		View:        serv.View,
		Subresource: app.SRI{},
		Version:     serv.Version,
		Brand:       serv.Brand,
		Environment: serv.Environment,
		RecordCount: serv.RecordCount,
	}
	dst, err := webapp.Templates(ctx, db)
	if err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if dst == nil {
		dst = make(map[string]*template.Template)
	}

	// copy HTML3 templates
	src, err := html3.Templates(ctx, sl, db, serv.View)
	if err != nil {
		return nil, fmt.Errorf(format, err)
	}
	maps.Copy(dst, src)

	// copy HTMX templates
	src = htmx.Templates(serv.View)
	maps.Copy(dst, src)

	return &TemplateRegistry{Templates: dst}, nil
}

// EchoConfig returns the base server start configuration.
func (serv *Server) EchoConfig() echo.StartConfig {
	config := echo.StartConfig{
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
	config echo.StartConfig, cert, key []byte, err error,
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
	config echo.StartConfig, cert, key []byte, err error,
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
	tlsConfig := echo.StartConfig{}
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
		return net.JoinHostPort(string(host), p)
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
