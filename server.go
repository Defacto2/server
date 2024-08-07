package main

/*
Package main is the entry point for the Defacto2 server application.

Use the Task runner / build tool (https://taskfile.dev) to build or run the source code.
$ task --list

Repository: 	https://github.com/Defacto2/server
Website:		https://defacto2.net
License:

© Defacto2, 2024
*/

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/caarlos0/env/v11"
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

var (
	ErrLog     = errors.New("cannot save logs")
	ErrPointer = errors.New("pointer is nil")
	ErrVer     = errors.New("postgresql version request failed")
)

// Main is the entry point for the application.
// By default the web server runs when no arguments are provided,
// otherwise, the command-line arguments are parsed and the application exits.
func main() {
	const exit = 0

	// initialize a temporary logger, get and print the environment variable configurations.
	logger, configs := environmentVars()
	if exitCode := parseFlags(logger, configs); exitCode >= exit {
		os.Exit(exitCode)
	}
	var w io.Writer = os.Stdout
	if configs.Quiet {
		w = io.Discard
	}
	fmt.Fprintf(w, "%s\n", configs)

	// connect to the database and perform some repairs and sanity checks.
	// if the database is cannot connect, the web server will continue.
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		logger.Errorf("main could not initialize the database data: %s", err)
	}
	defer db.Close()
	ctx := context.WithValue(context.Background(), helper.LoggerKey, logger)
	var database postgres.Version
	if err := database.Query(db); err != nil {
		logger.Errorf("postgres version query: %w", err)
	}
	if db != nil && tx != nil {
		repairs(ctx, db, tx, configs)
	}
	sanityChecks(ctx, configs)

	// start the web server and the sugared logger.
	website := newInstance(ctx, db, configs)
	logger = serverLog(configs, website.RecordCount)
	router := website.Controller(logger)
	website.Info(logger, w)
	if err := website.Start(router, logger, configs); err != nil {
		logger.Fatalf("%s: please check the environment variables", err)
	}

	go func() {
		// get the owner and group of the current process and print them to the console.
		if groups, usr, err := helper.Owner(); err != nil {
			logger.Errorf("owner in main: %s", err)
		} else {
			clean := slices.DeleteFunc(groups, func(e string) bool {
				return e == ""
			})
			fmt.Fprintf(w, "Running as %s for the groups, %s.\n", usr, strings.Join(clean, ","))
		}
		// get the local IP addresses and print them to the console.
		if localIPs, err := configs.Addresses(); err != nil {
			logger.Errorf("configs addresses in main: %s", err)
		} else {
			fmt.Fprintf(w, "%s\n", localIPs)
		}
	}()

	// shutdown the web server.
	website.ShutdownHTTP(router, logger)
}

// environmentVars is used to parse the environment variables and set the Go runtime.
// Defaults are used if the environment variables are not set.
func environmentVars() (*zap.SugaredLogger, config.Config) {
	logger := zaplog.Status().Sugar()
	configs := config.Config{
		Compression:   true,
		DatabaseURL:   postgres.DefaultURL,
		HTTPPort:      config.HTTPPort,
		ProdMode:      true,
		ReadOnly:      true,
		SessionMaxAge: config.SessionHours,
	}
	if err := env.Parse(&configs); err != nil {
		logger.Fatalf("could not parse the environment variable, it probably contains an invalid value: %s", err)
	}
	configs.Override()

	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}
	return logger, configs
}

// newInstance is used to create the server controller instance.
func newInstance(ctx context.Context, db *sql.DB, configs config.Config) handler.Configuration {
	c := handler.Configuration{
		Brand:       brand,
		Environment: configs,
		Public:      public,
		Version:     version,
		View:        view,
	}
	if c.Version == "" {
		c.Version = cmd.Commit("")
	}
	if ctx != nil && db != nil {
		c.RecordCount = recordCount(ctx, db)
	}
	return c
}

// parseFlags is used to parse the commandline arguments.
// If an error is returned, the application will exit with the error code.
// Otherwise, a negative value is returned to indicate the application should continue.
func parseFlags(logger *zap.SugaredLogger, configs config.Config) int {
	if logger == nil {
		return -1
	}
	code, err := cmd.Run(version, &configs)
	if err != nil {
		logger.Errorf("run command, parse flags: %s", err)
		return int(code)
	}
	useExitCode := code >= cmd.ExitOK
	if useExitCode {
		return int(code)
	}
	return -1
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the Production mode environment variable is set.to false.
func sanityChecks(ctx context.Context, configs config.Config) {
	if !configs.ProdMode {
		return
	}
	logger := helper.Logger(ctx)
	if err := configs.Checks(logger); err != nil {
		logger.Errorf("sanity checks could not read the environment variable, it probably contains an invalid value: %s", err)
	}
	checks(logger, configs.ReadOnly)
	conn, err := postgres.New()
	if err != nil {
		logger.Errorf("sanity checks could not initialize the database data: %s", err)
		return
	}
	_ = conn.Validate(logger)
}

// checks is used to confirm the required commands are available.
// These are skipped if readonly is true.
func checks(logger *zap.SugaredLogger, readonly bool) {
	if logger == nil || readonly {
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
		logger.Warnln("The following commands are required for the server to run in WRITE MODE",
			"\n\t\t\tThese need to be installed and accessible on the system path:"+
				"\t\t\t"+buf.String())
	}
	if err := command.LookupUnrar(); err != nil {
		if errors.Is(err, command.ErrVers) {
			logger.Warnf("Found unrar but " +
				"could not find unrar by Alexander Roshal, " +
				"is unrar-free mistakenly installed?")
			return
		}
		logger.Warnf("lookup unrar check: %s", err)
	}
}

// repairs is used to fix any known issues with the file assets and the database entries.
// These are skipped if the Production mode environment variable is set to false.
func repairs(ctx context.Context, db *sql.DB, tx *sql.Tx, configs config.Config) {
	if !configs.ProdMode {
		return
	}
	logger := helper.Logger(ctx)
	if db == nil || tx == nil {
		logger.Errorf("repairs is missing a required parameter")
		return
	}
	if err := configs.RepairAssets(ctx, db); err != nil {
		logger.Errorf("asset repairs: %s", err)
	}
	if err := repairDatabase(ctx, db, tx); err != nil {
		if errors.Is(err, ErrVer) {
			logger.Warnf("A %s, is the database server down?", ErrVer)
		} else {
			logger.Errorf("repair database could not initialize the database data: %s", err)
		}
	}
}

// serverLog is used to setup the logger for the server and print the startup message.
func serverLog(configs config.Config, count int) *zap.SugaredLogger {
	logger := zaplog.Timestamp().Sugar()
	const welcome = "Welcome to the Defacto2 web application"
	switch {
	case count == 0:
		logger.Error(welcome + " with no database records")
	case config.MinimumFiles > count:
		logger.Warnf(welcome+" with only %d records, expecting at least %d+", count, config.MinimumFiles)
	default:
		logger.Infof(welcome+" containing %d records", count)
	}
	if configs.ProdMode {
		if err := configs.LogStore(); err != nil {
			logger.Fatalf("%w using server log: %s", ErrLog, err)
		}
		logger = zaplog.Store(configs.AbsLog).Sugar()
	}
	return logger
}

// repairDatabase on startup checks the database connection and make any data corrections.
func repairDatabase(ctx context.Context, db *sql.DB, tx *sql.Tx) error {
	if db == nil || tx == nil {
		return fmt.Errorf("%w: %s", ErrPointer,
			"the repair database is missing a required parameter")
	}
	logger := helper.Logger(ctx)
	if err := fix.Artifacts.Run(ctx, db, tx); err != nil {
		defer func() {
			if err := tx.Rollback(); err != nil {
				logger.Error(err)
			}
		}()
		return fmt.Errorf("repair database could not fix all artifacts: %w", err)
	}
	return nil
}

// recordCount returns the number of records in the database.
func recordCount(ctx context.Context, db *sql.DB) int {
	if db == nil {
		return 0
	}
	fs, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}
	return int(fs)
}
