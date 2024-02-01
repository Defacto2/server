package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logger"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/caarlos0/env/v10"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

//go:embed public/text/defacto2.txt
var brand []byte

//go:embed public/**/*
var public embed.FS

//go:embed view/**/*
var view embed.FS

// version is generated by the GoReleaser ldflags.
var version string

// LocalMode is used to always override the PRODUCTION_MODE and READ_ONLY environment variables.
// It removes the option to set a number environment variables when running the server locally.
// This is set using the -ldflags option when building the app.
//
// Example, go build -ldflags="-X 'main.LocalMode=true'", this will set the LocalMode variable to true.
var LocalMode string //nolint:gochecknoglobals

var (
	ErrCmd = errors.New("the command given did not work")
	ErrDB  = errors.New("could not initialize the database data")
	ErrEnv = errors.New("environment variable probably contains an invalid value")
	ErrFS  = errors.New("the directories repair broke")
	ErrLog = errors.New("the server cannot save any logs")
	ErrVer = errors.New("postgresql version request failed")
)

// main is the entry point for the application.
func main() { //nolint:funlen
	// Logger
	// Use the development log until the environment vars are parsed
	logs := logger.CLI().Sugar()

	// Environment variables configuration
	configs := config.Config{}
	if err := env.ParseWithOptions(
		&configs, env.Options{Prefix: config.EnvPrefix}); err != nil {
		logs.Fatalf("%w: %s", ErrEnv, err)
	}
	configs.Override(localMode())

	// Go runtime customizations
	// If not set, the automaxprocs lib automatically set GOMAXPROCS to match Linux container CPU quota
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// By default the web server runs when no arguments are provided
	commandLine(logs, configs)

	// Configuration sanity checks
	if err := configs.Checks(logs); err != nil {
		logs.Errorf("%s: %s", ErrEnv, err)
	}

	// Confirm command requirements when not running in read-only mode
	checks(logs, configs.ReadMode)

	// Database connection checks
	if conn, err := postgres.New(); err != nil {
		logs.Errorf("%s: %s", ErrDB, err)
	} else {
		_ = conn.Check(logs, localMode())
	}

	// Repair assets on the host file system
	if err := configs.RepairFS(logs); err != nil {
		logs.Errorf("%s: %s", ErrFS, err)
	}

	// Setup the logger and print the startup production/read-only message
	logs = setupLogger(logs, configs)

	// Echo router and controller instance
	server := handler.Configuration{
		Brand:   &brand,
		Import:  &configs,
		Logger:  logs,
		Public:  public,
		Version: version,
		View:    view,
	}
	if server.Version == "" {
		server.Version = cmd.Commit("")
	}

	// Repair the database on startup
	if err := RepairDB(); err != nil {
		repairdb(logs, err)
	}
	server.RecordCount = RecordCount()

	// Controllers and routes
	e := server.Controller()

	// Startup information and warnings
	server.Info()

	// Start the HTTP and the TLS server
	switch {
	case configs.UseTLS() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			server.StartHTTP(e2)
		}()
		go server.StartTLS(e)
	case configs.UseTLSLocal() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			server.StartHTTP(e2)
		}()
		go server.StartTLSLocal(e)
	case configs.UseTLS():
		go server.StartTLS(e)
	case configs.UseHTTP():
		go server.StartHTTP(e)
	case configs.UseTLSLocal():
		go server.StartTLSLocal(e)
	default:
		// this should never happen as HTTPPort is always set to a default value
		logs.Fatalf("No server ports are configured, please check the environment variables.")
	}

	// List the local IP addresses that can also be used to access the server
	go func() {
		s, err := configs.Startup()
		if err != nil {
			logs.Errorf("%s: %s", ErrEnv, err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", s)
	}()
	if localMode() {
		go func() {
			fmt.Fprint(os.Stdout, "Tap Ctrl + C, to exit at anytime.\n")
		}()
	}
	// Gracefully shutdown the HTTP server
	server.ShutdownHTTP(e)
}

// commandLine is used to parse the command-line arguments.
func commandLine(logs *zap.SugaredLogger, c config.Config) {
	if logs == nil {
		return
	}
	code, err := cmd.Run(version, &c)
	if err != nil {
		logs.Errorf("%s: %s", ErrCmd, err)
		os.Exit(int(code))
	}
	useExitCode := code >= cmd.ExitOK
	if useExitCode {
		os.Exit(int(code))
	}
	// after the command-line arguments are parsed, continue with the web server
}

// setupLogger is used to setup the logger.
func setupLogger(logs *zap.SugaredLogger, c config.Config) *zap.SugaredLogger {
	if logs == nil {
		return nil
	}
	if localMode() {
		s := "Welcome to the local Defacto2 web application."
		logs.Info(s)
		return logs
	}
	mode := "read-only mode"
	if !c.ReadMode {
		mode = "write mode"
	}
	switch c.ProductionMode {
	case true:
		if err := c.LogStorage(); err != nil {
			logs.Fatalf("%w: %s", ErrLog, err)
		}
		logs = logger.Production(c.LogDir).Sugar()
		s := "The server is running in a "
		s += strings.ToUpper("production, "+mode) + "."
		logs.Info(s)
	default:
		s := "The server is running in a "
		s += strings.ToUpper("development, "+mode) + "."
		logs.Warn(s)
		logs = logger.Development().Sugar()
	}
	return logs
}

// checks is used to confirm the required commands are available.
func checks(logs *zap.SugaredLogger, readonly bool) {
	if logs == nil || readonly {
		return
	}
	var buf strings.Builder
	for i, name := range command.Lookups() {
		if err := command.LookCmd(name); err != nil {
			buf.WriteString("\n\t\t\tmissing: " + name)
			buf.WriteString("\t" + command.Infos()[i])
		}
	}
	if buf.Len() > 0 {
		logs.Warnln("The following commands are required for the server to run in WRITE MODE",
			"\n\t\t\tThese need to be installed and accessible on the system path:"+
				"\t\t\t"+buf.String())
	}
	if err := command.LookupUnrar(); err != nil {
		if errors.Is(err, command.ErrVers) {
			logs.Warnf("Found unrar but " +
				"could not find unrar by Alexander Roshal, " +
				"is unrar-free mistakenly installed?")
		} else {
			logs.Warnf("%s: %s", ErrCmd, err)
		}
	}
}

// localMode is used to always override the PRODUCTION_MODE and READ_ONLY environment variables.
func localMode() bool {
	val, err := strconv.ParseBool(LocalMode)
	if err != nil {
		return false
	}
	return val
}

// RepairDB on startup checks the database connection and make any data corrections.
func RepairDB() error {
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	var ver postgres.Version
	ctx := context.Background()
	if err := ver.Query(); err != nil {
		return ErrVer
	}
	if localMode() {
		return nil
	}
	return fix.Releaser.Run(ctx, os.Stderr, db)
}

// repairdb is used to log the database repair error.
func repairdb(z *zap.SugaredLogger, err error) {
	if z == nil || err == nil {
		return
	}
	if errors.Is(err, ErrVer) {
		z.Warnf("A %s, is the database server down?", ErrVer)
	} else {
		z.Errorf("%s: %s", ErrDB, err)
	}
}

// RecordCount returns the number of records in the database.
func RecordCount() int {
	db, err := postgres.ConnectDB()
	if err != nil {
		return 0
	}
	defer db.Close()
	ctx := context.Background()
	fs, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}
	return int(fs)
}
