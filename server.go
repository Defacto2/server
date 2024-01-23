package main

import (
	"context"
	"crypto/sha512"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logger"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
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

var (
	ErrCmd = errors.New("the command given did not work")
	ErrDB  = errors.New("could not initialize the database data")
	ErrEnv = errors.New("environment variable probably contains an invalid value")
	ErrFS  = errors.New("the directories repair broke")
	ErrLog = errors.New("the server cannot save any logs")
	ErrVer = errors.New("postgresql version request failed")
	ErrZap = errors.New("the logger instance is nil")
)

const (
	uuid = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

func main() {
	// Logger
	// Use the development log until the environment vars are parsed
	logs := logger.CLI().Sugar()

	// Environment variables configuration
	configs := config.Config{}
	if err := env.ParseWithOptions(
		&configs, env.Options{Prefix: config.EnvPrefix}); err != nil {
		logs.Fatalf("%w: %s", ErrEnv, err)
	}
	configs = *Override(&configs)

	// Go runtime customizations
	// If not set, the automaxprocs lib automatically set GOMAXPROCS to match Linux container CPU quota
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// By default the web server runs when no arguments are provided
	commandLine(logs, configs)

	// Configuration sanity checks
	configs.Checks(logs)

	// Confirm command requirements when not running in read-only mode
	checks(logs, configs.ReadMode)

	// Database connection checks
	if conn, err := postgres.New(); err != nil {
		logs.Errorf("%s: %s", ErrDB, err)
	} else {
		_ = conn.Check(logs)
	}

	// Repair assets on the host file system
	if err := RepairFS(logs, &configs); err != nil {
		logs.Errorf("%s: %s", ErrFS, err)
	}

	// Setup the logger and print the startup production/read-only message
	logs = initLogger(logs, configs)

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

	// Start the HTTP and HTTPS server
	go server.StartHTTP(e)
	go server.StartHTTPS(e)

	// List the local IP addresses that can also be used to access the server
	go fmt.Fprintf(os.Stdout, "%s\n", configs.Startup())

	// Gracefully shutdown the HTTP server
	server.ShutdownHTTP(e)
}

func commandLine(logs *zap.SugaredLogger, configs config.Config) {
	const exitProgram = 0
	if code, err := cmd.Run(version, &configs); err != nil {
		logs.Errorf("%s: %s", ErrCmd, err)
		os.Exit(code)
	} else if code >= exitProgram {
		os.Exit(code)
	}
}

func initLogger(logs *zap.SugaredLogger, configs config.Config) *zap.SugaredLogger {
	// Setup the logger
	mode := "read-only mode"
	if !configs.ReadMode {
		mode = "write mode"
	}
	switch configs.ProductionMode {
	case true:
		if err := configs.LogStorage(); err != nil {
			logs.Fatalf("%w: %s", ErrLog, err)
		}
		logs = logger.Production(configs.LogDir).Sugar()
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

func checks(logs *zap.SugaredLogger, isReadOnly bool) {
	if isReadOnly {
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
			logs.Warnf("Could not find unrar by Alexander Roshal, " +
				"is the unrar-free command mistakenly installed?")
		} else {
			logs.Warnf("%s: %s", ErrCmd, err)
		}
	}
}

// Override the configuration settings fetched from the environment.
func Override(c *config.Config) *config.Config {
	// hash and delete any supplied google ids
	ids := strings.Split(c.GoogleIDs, ",")
	for _, id := range ids {
		sum := sha512.Sum384([]byte(id))
		c.GoogleAccounts = append(c.GoogleAccounts, sum)
	}
	c.GoogleIDs = "overwrite placeholder"
	c.GoogleIDs = "" // empty the string

	if c.HTTPPort == 0 && c.HTTPSPort == 0 {
		c.HTTPPort = 1323
	}

	// examples of hard-coded overrides:
	// c.ProductionMode = true
	// c.ReadMode = true
	return c
}

// RepairDB, on startup check the database connection and make any data corrections.
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
	return model.RepairReleasers(ctx, os.Stderr, db)
}

func repairdb(z *zap.SugaredLogger, err error) {
	if errors.Is(err, ErrVer) {
		z.Warnf("%s, is the database server down?", ErrVer)
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
	x, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}
	return int(x)
}

// RepairFS, on startup check the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning.
func RepairFS(z *zap.SugaredLogger, c *config.Config) error {
	if z == nil {
		return ErrZap
	}
	dirs := []string{c.PreviewDir, c.ThumbnailDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		z.Info("scan: ", dir)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			name := info.Name()
			if info.IsDir() {
				return fixDir(name, path, dir)
			}
			return fixImgs(name, path)
		})
		if err != nil {
			return err
		}
	}
	dir := c.DownloadDir
	if _, err := os.Stat(dir); err != nil {
		var ignore error
		return ignore //nolint:nilerr
	}
	z.Info("scan: ", dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if info.IsDir() {
			return fixDir(name, path, dir)
		}
		return fixDL(name, path)
	})
}

func rm(name, info, path string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", info, name)
	defer os.Remove(path)
}

func fixDir(name, path, dir string) error {
	const st = ".stfolder" // st is a syncthing directory
	switch name {
	case filepath.Base(dir):
		// skip the root directory
	case st:
		defer os.RemoveAll(path)
	default:
		fmt.Fprintln(os.Stderr, "unknown dir:", path)
	}
	return nil // always skip
}

func fixDL(name, path string) error {
	l := len(name)
	switch filepath.Ext(name) {
	case ".chiptune", ".txt":
		return nil
	case ".zip":
		if l != len(uuid)+4 && l != len(cfid)+4 {
			rm(name, "remove", path)
		}
		return nil
	default:
		if l != len(uuid) && l != len(cfid) {
			rm(name, "unknown", path)
		}
	}
	return nil
}

func fixImgs(name, path string) error {
	const (
		png  = ".png"    // png file extension
		webp = ".webp"   // webp file extension
		lpng = len(png)  // length of png file extension
		lweb = len(webp) // length of webp file extension
	)
	ext := filepath.Ext(name)
	l := len(name)
	switch ext {
	case png:
		if l != len(uuid)+lpng && l != len(cfid)+lpng {
			rm(name, "remove", path)
		}
	case webp:
		if l != len(uuid)+lweb && l != len(cfid)+lweb {
			rm(name, "remove", path)
		}
	default:
		rm(name, "unknown", path)
	}
	return nil
}
