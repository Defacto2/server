package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
)

type checkStatus int

const (
	statusOK checkStatus = iota
	statusSkip
	statusWarn
	statusFail
)

type checkItem struct {
	status checkStatus
	name   string
	detail string
}

type checkGroup struct {
	title string
	items []checkItem
}

// Checks runs sanity checks for the environment variable configurations
// and writes a grouped report to w. Returns ErrCheckFailed if any check fails.
func (c *Config) Checks(ctx context.Context, w io.Writer) error {
	const format = "config checks %s: %w"
	if err := nils.Check(ctx); err != nil {
		return fmt.Errorf(format, "arguments", err)
	}
	if w == nil {
		w = os.Stdout
	}

	color := logs.Color(w)
	style := func(code, s string) string {
		if !color || s == "" {
			return s
		}
		return code + s + ansiReset
	}

	groups := []checkGroup{
		c.checkDirs(),
		c.checkNetwork(),
		c.checkDatabase(),
		c.checkModeAuth(),
		c.checkTools(ctx),
	}

	var failed, warned int
	_, _ = fmt.Fprintln(w, style(ansiBold, "Web Server Configuration Checks"))

	for _, group := range groups {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, style(ansiBold, group.title))
		for _, item := range group.items {
			var badge string
			switch item.status {
			case statusOK:
				badge = style(ansiGreen, "OK  ")
			case statusSkip:
				badge = style(ansiDim, "SKIP")
			case statusWarn:
				warned++
				badge = style(ansiYellow, "WARN")
			case statusFail:
				failed++
				badge = style(ansiBoldRed, "FAIL")
			}

			name := fmt.Sprintf("%-*s", colWidth, item.name)
			detail := item.detail
			if item.status == statusSkip {
				detail = style(ansiDim, detail)
			}
			_, _ = fmt.Fprintf(w, "  %s  %s %s\n", badge, style(ansiCyan, name), detail)
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, formatCheckSummary(failed, warned, style))

	if failed > 0 {
		return fmt.Errorf("%w (%d found)", ErrCheckFailed, failed)
	}
	return nil
}

func formatCheckSummary(failed, warned int, style func(string, string) string) string {
	var parts []string
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if warned > 0 {
		label := "warnings"
		if warned == 1 {
			label = "warning"
		}
		parts = append(parts, fmt.Sprintf("%d %s", warned, label))
	}

	switch {
	case failed > 0:
		return style(ansiBoldRed, "Result: "+strings.Join(parts, ", "))
	case warned > 0:
		return style(ansiYellow, "Result: all checks passed ("+strings.Join(parts, ", ")+")")
	default:
		return style(ansiGreen, "Result: all checks passed")
	}
}

func (c *Config) checkDirs() checkGroup {
	items := []checkItem{
		checkAssetDir("D2_DIR_DOWNLOAD", Directory(c.AbsDownload)),
		checkAssetDir("D2_DIR_PREVIEW", Directory(c.AbsPreview)),
		checkAssetDir("D2_DIR_THUMBNAIL", Directory(c.AbsThumbnail)),
		checkAssetDir("D2_DIR_EXTRA", Directory(c.AbsExtra)),
		checkAssetDir("D2_DIR_ORPHANED", Directory(c.AbsOrphaned)),
		c.checkLogDirItem(),
	}
	return checkGroup{title: "DIRECTORIES", items: items}
}

func checkAssetDir(env string, d Directory) checkItem {
	if d == "" {
		return checkItem{status: statusFail, name: env, detail: "not configured"}
	}
	if issue := d.Issue(); issue != "" {
		return checkItem{status: statusFail, name: env, detail: fmt.Sprintf("%s — %s", d, issue)}
	}
	return checkItem{status: statusOK, name: env, detail: d.String()}
}

func (c *Config) checkLogDirItem() checkItem {
	const env = "D2_DIR_LOG"
	if c.AbsLog == "" {
		return checkItem{status: statusSkip, name: env, detail: "not set — file logging disabled"}
	}
	if err := c.checkLogDir(); err != nil {
		return checkItem{status: statusFail, name: env, detail: fmt.Sprintf("%s — %v", c.AbsLog, err)}
	}
	return checkItem{status: statusOK, name: env, detail: fmt.Sprintf("%s (writable)", c.AbsLog)}
}

func (c *Config) checkNetwork() checkGroup {
	items := []checkItem{
		checkPort("D2_HTTP_PORT", Port(c.HTTPPort)),
		checkPort("D2_TLS_PORT", Port(c.TLSPort)),
		checkTLSFile("D2_TLS_CERT", File(c.TLSCert), c.TLSKey != "", "D2_TLS_KEY"),
		checkTLSFile("D2_TLS_KEY", File(c.TLSKey), c.TLSCert != "", "D2_TLS_CERT"),
	}
	return checkGroup{title: "NETWORK & TLS", items: items}
}

func checkPort(env string, p Port) checkItem {
	if p == 0 {
		return checkItem{status: statusSkip, name: env, detail: "disabled (0)"}
	}
	portStr := strconv.FormatUint(uint64(p), 10)
	if err := p.Check(); err != nil {
		return checkItem{status: statusFail, name: env, detail: fmt.Sprintf("%s — %v", portStr, err)}
	}
	return checkItem{status: statusOK, name: env, detail: portStr}
}

func checkTLSFile(env string, f File, pairSet bool, pairEnv string) checkItem {
	if f == "" {
		if pairSet {
			return checkItem{status: statusFail, name: env, detail: "not set — required when " + pairEnv + " is set"}
		}
		return checkItem{status: statusSkip, name: env, detail: "not set"}
	}
	if issue := f.Issue(); issue != "" {
		return checkItem{status: statusFail, name: env, detail: fmt.Sprintf("%s — %s", f, issue)}
	}
	return checkItem{status: statusOK, name: env, detail: f.String()}
}

func (c *Config) checkDatabase() checkGroup {
	const env = "D2_DATABASE_URL"
	rawURL := string(c.DatabaseURL)
	if rawURL == "" {
		return checkGroup{
			title: "DATABASE",
			items: []checkItem{{status: statusWarn, name: env, detail: "not set"}},
		}
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		// Never include err.Error(), as url.Parse echoes the raw URL including credentials.
		return checkGroup{
			title: "DATABASE",
			items: []checkItem{{status: statusWarn, name: env, detail: mask + " — database connection URL is invalid"}},
		}
	}
	if u.Scheme != postgres.Protocol && u.Scheme != "postgresql" {
		detail := fmt.Sprintf("%s — unexpected scheme %q (expected postgres or postgresql)",
			c.DatabaseURL.String(), u.Scheme)
		return checkGroup{
			title: "DATABASE",
			items: []checkItem{{status: statusWarn, name: env, detail: detail}},
		}
	}

	return checkGroup{
		title: "DATABASE",
		items: []checkItem{{status: statusOK, name: env, detail: c.DatabaseURL.String()}},
	}
}

func (c *Config) checkModeAuth() checkGroup {
	items := []checkItem{
		{status: statusOK, name: "D2_PROD_MODE", detail: boolLabel(c.ProdMode)},
		checkToggleWarn("D2_READ_ONLY", c.ReadOnly,
			"on — edits to the database are not allowed",
			"off — write mode enabled"),
		checkToggleWarn("D2_NO_CRAWL", c.NoCrawl,
			"on — search engine crawling is disabled",
			"off"),
	}
	if c.ProdMode.Bool() && !c.ReadOnly.Bool() {
		items = append(items, c.checkProdAuth()...)
	}
	items = append(items, c.checkDemozooKey())
	return checkGroup{title: "MODE & AUTH", items: items}
}

func checkToggleWarn(env string, t Toggle, onDetail, offDetail string) checkItem {
	if t.Bool() {
		return checkItem{status: statusWarn, name: env, detail: onDetail}
	}
	return checkItem{status: statusOK, name: env, detail: offDetail}
}

func (c *Config) checkProdAuth() []checkItem {
	clientID := checkItem{status: statusOK, name: "D2_GOOGLE_CLIENT_ID", detail: mask}
	if c.GoogleClientID == "" {
		clientID.status = statusWarn
		clientID.detail = "not set — required for admin sign-in"
	}

	googleIDs := checkItem{status: statusOK, name: "D2_GOOGLE_IDS", detail: mask}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		googleIDs.status = statusWarn
		googleIDs.detail = "not set — no admin accounts configured"
	} else if s := c.GoogleAccounts.String(); s != "" {
		googleIDs.detail = fmt.Sprintf("%s (%s)", mask, s)
	}

	sessionAge := checkItem{status: statusOK, name: "D2_SESSION_MAX_AGE", detail: c.SessionMaxAge.String()}
	if c.SessionMaxAge == 0 {
		sessionAge.status = statusWarn
		sessionAge.detail = "0 — admin sessions never expire"
	}

	return []checkItem{clientID, googleIDs, sessionAge}
}

func (c *Config) checkDemozooKey() checkItem {
	const env = "D2_DEMOZOO_KEY"
	switch {
	case c.DemozooAuth == "":
		return checkItem{status: statusSkip, name: env, detail: "not set"}
	case c.DemozooAuth.Issue() != "":
		return checkItem{
			status: statusFail,
			name:   env,
			detail: fmt.Sprintf("%s — %s", mask, c.DemozooAuth.Issue()),
		}
	default:
		return checkItem{status: statusOK, name: env, detail: mask}
	}
}

func (c *Config) checkTools(ctx context.Context) checkGroup {
	const groupTitle = "EXTERNAL TOOLS"
	if c.ReadOnly.Bool() {
		return checkGroup{
			title: groupTitle,
			items: []checkItem{{
				status: statusSkip,
				name:   "Write-mode tools",
				detail: "skipped in read-only mode",
			}},
		}
	}

	var items []checkItem
	hasUnrar := true
	for i, file := range command.Lookups {
		if _, err := command.Lookup(file); err != nil {
			if file == command.Unrar {
				hasUnrar = false
			}
			items = append(items, checkItem{
				status: statusWarn,
				name:   file,
				detail: "missing from PATH — " + command.Infos[i],
			})
		}
	}

	if hasUnrar {
		if err := command.LookupUnrar(ctx); err != nil {
			detail := err.Error()
			if errors.Is(err, command.ErrVersion) {
				detail = "not authored by Alexander Roshal (is unrar-free installed?)"
			}
			items = append(items, checkItem{
				status: statusWarn,
				name:   command.Unrar,
				detail: detail,
			})
		}
	}

	if len(items) == 0 {
		items = []checkItem{{
			status: statusOK,
			name:   "Write-mode tools",
			detail: fmt.Sprintf("all %d tools available (unrar verified)", len(command.Lookups)),
		}}
	}

	return checkGroup{title: groupTitle, items: items}
}

func boolLabel(t Toggle) string {
	if t.Bool() {
		return "on"
	}
	return "off"
}

// checkLogDir runs checks against the configured log directory.
// If no log directory is configured, file logging is disabled and the check is skipped.
func (c *Config) checkLogDir() error {
	if c.AbsLog == "" {
		return nil
	}

	path := string(c.AbsLog)
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if !st.IsDir() {
		return ErrNotDir
	}

	// test write permissions by creating and removing a temp file directly
	name := filepath.Join(path, ".defacto2_touch_test")
	if _, sErr := os.Stat(name); sErr == nil {
		if rErr := os.Remove(name); rErr != nil {
			return fmt.Errorf("cannot remove test file: %w", rErr)
		}
	}

	f, err := os.Create(name)
	if err != nil {
		return errors.Join(err, ErrTouch)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("cannot close test file: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("cannot remove test file: %w", err)
	}
	return nil
}
