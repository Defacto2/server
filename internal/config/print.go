//nolint:exhaustruct_v5
package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
)

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiCyan    = "\033[36m"
	ansiBoldRed = "\033[1;31m"

	unsetLabel = "(unset)"
	colWidth   = 21
)

type printItem struct {
	env    string
	value  string
	muted  bool
	active bool
	tip    string
	issue  string
}

type printGroup struct {
	title string
	items []printItem
}

// Print writes a grouped, human-readable summary of the server configuration to w.
// When detail is true, each setting includes its full description from Config.Help().
// ANSI styling is enabled automatically when w is a terminal.
func (c Config) Print(w io.Writer, detail bool) {
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

	help := c.Help()
	indent := strings.Repeat(" ", colWidth+3)

	_, _ = fmt.Fprintln(w, style(ansiBold, "Web Server Configurations"))

	for _, group := range c.groups() {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, style(ansiBold, group.title))
		for _, item := range group.items {
			name := fmt.Sprintf("%-*s", colWidth, item.env)
			val := item.value
			switch {
			case item.muted:
				val = style(ansiDim, val)
			case item.active:
				val = style(ansiGreen, val)
			}

			var note string
			switch {
			case item.issue != "":
				note = "  " + style(ansiBoldRed, "ISSUE: "+item.issue)
			case item.tip != "":
				note = " " + style(ansiDim, "— "+item.tip)
			}

			_, _ = fmt.Fprintf(w, "  %s %s%s\n", style(ansiCyan, name), val, note)
			if detail {
				if desc := help[item.env]; desc != "" {
					_, _ = fmt.Fprintf(w, "%s%s\n", indent, style(ansiDim, desc))
				}
			}
		}
	}
}

func (c Config) groups() []printGroup {
	return []printGroup{
		{
			title: "DIRECTORIES",
			items: []printItem{
				pathItem("D2_DIR_LOG", c.AbsLog.String(), c.AbsLog.Help(), c.AbsLog.Issue()),
				pathItem("D2_DIR_DOWNLOAD", c.AbsDownload.String(), c.AbsDownload.Help(), c.AbsDownload.Issue()),
				pathItem("D2_DIR_PREVIEW", c.AbsPreview.String(), c.AbsPreview.Help(), c.AbsPreview.Issue()),
				pathItem("D2_DIR_THUMBNAIL", c.AbsThumbnail.String(), c.AbsThumbnail.Help(), c.AbsThumbnail.Issue()),
				pathItem("D2_DIR_EXTRA", c.AbsExtra.String(), c.AbsExtra.Help(), c.AbsExtra.Issue()),
				pathItem("D2_DIR_ORPHANED", c.AbsOrphaned.String(), c.AbsOrphaned.Help(), c.AbsOrphaned.Issue()),
			},
		},
		{
			title: "DATABASE & AUTH",
			items: []printItem{
				pathItem("D2_DATABASE_URL", c.DatabaseURL.String(), "", ""),
				secretItem("D2_SESSION_KEY", c.SessionKey.String(), c.SessionKey.Help(), ""),
				{
					env:   "D2_SESSION_MAX_AGE",
					value: c.SessionMaxAge.String(),
				},
				secretItem("D2_GOOGLE_CLIENT_ID", c.GoogleClientID.String(), c.GoogleClientID.Help(), ""),
				c.google(),
				secretItem("D2_DEMOZOO_KEY", c.DemozooAuth.String(), c.DemozooAuth.Help(), c.DemozooAuth.Issue()),
			},
		},
		{
			title: "NETWORK & TLS",
			items: []printItem{
				portItem("D2_HTTP_PORT", uint16(c.HTTPPort), c.HTTPPort.Help()),
				portItem("D2_TLS_PORT", uint16(c.TLSPort), c.TLSPort.Help()),
				pathItem("D2_TLS_CERT", c.TLSCert.String(), c.TLSCert.Help(), c.TLSCert.Issue()),
				pathItem("D2_TLS_KEY", c.TLSKey.String(), c.TLSKey.Help(), c.TLSKey.Issue()),
				pathItem("D2_MATCH_HOST", c.MatchHost.String(), c.MatchHost.Help(), ""),
			},
		},
		{
			title: "RUNTIME",
			items: []printItem{
				{
					env:   "D2_MAX_PROCS",
					value: c.MaxProcs.String(),
					muted: c.MaxProcs == 0,
					tip:   c.MaxProcs.Help(),
				},
				toggleItem("D2_QUIET", c.Quiet),
				toggleItem("D2_COMPRESSION", c.Compression),
				toggleItem("D2_PROD_MODE", c.ProdMode),
				toggleItem("D2_READ_ONLY", c.ReadOnly),
				toggleItem("D2_NO_CRAWL", c.NoCrawl),
				toggleItem("D2_LOG_ALL", c.LogAll),
			},
		},
	}
}

func pathItem(env, value, tip, issue string) printItem {
	if value == "" {
		return printItem{env: env, value: unsetLabel, muted: true, tip: tip, issue: issue}
	}
	return printItem{env: env, value: value, tip: tip, issue: issue}
}

func secretItem(env, value, tip, issue string) printItem {
	if value == "" {
		return printItem{env: env, value: unsetLabel, muted: true, tip: tip, issue: issue}
	}
	return printItem{env: env, value: mask, tip: tip, issue: issue}
}

func (c Config) google() printItem {
	if len(c.GoogleAccounts) == 0 {
		return printItem{
			env:   "D2_GOOGLE_IDS",
			value: unsetLabel,
			muted: true,
			tip:   c.GoogleAccounts.Help(),
		}
	}
	return printItem{
		env:   "D2_GOOGLE_IDS",
		value: fmt.Sprintf("%s (%s)", mask, c.GoogleAccounts.String()),
		tip:   c.GoogleAccounts.Help(),
	}
}

func portItem(env string, port uint16, tip string) printItem {
	return printItem{
		env:   env,
		value: strconv.FormatUint(uint64(port), 10),
		muted: port == 0,
		tip:   tip,
	}
}

func toggleItem(env string, t Toggle) printItem {
	if t.Bool() {
		return printItem{env: env, value: "on", active: true}
	}
	return printItem{env: env, value: "off", muted: true}
}

// AddressGrouped writes a grouped, human-readable list of server URLs to w.
// ANSI styling is enabled automatically when w is a terminal.
func (c Config) AddressGrouped(w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}
	httpURLs, httpsURLs, err := c.addressURLs()
	if err != nil {
		return fmt.Errorf("config addresses: %w", err)
	}

	color := logs.Color(w)
	style := func(code, s string) string {
		if !color || s == "" {
			return s
		}
		return code + s + ansiReset
	}

	_, _ = fmt.Fprintln(w, style(ansiBold, "Web Server Addresses"))
	_, _ = fmt.Fprintln(w)
	if host := c.MatchHost.String(); host != "" {
		_, _ = fmt.Fprintf(w, "%s %s\n", style(ansiDim, "Bound to host:"), host)
	} else {
		_, _ = fmt.Fprintln(w, style(ansiDim,
			"Depending on the firewall and operating system setup, the server may be accessible at:"))
	}

	if len(httpURLs) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, style(ansiBold, fmt.Sprintf("HTTP (:%d)", c.HTTPPort)))
		for _, link := range httpURLs {
			_, _ = fmt.Fprintf(w, "  %s\n", style(ansiCyan, link))
		}
	}

	if len(httpsURLs) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, style(ansiBold, fmt.Sprintf("HTTPS (:%d)", c.TLSPort)))
		for _, link := range httpsURLs {
			_, _ = fmt.Fprintf(w, "  %s\n", style(ansiCyan, link))
		}
	}

	return nil
}

// PurgeTemp purges the application temporary directories and
// writes a summary report to w.
// When dryrun is true a preview is run.
func (c Config) PurgeTemp(w io.Writer, dryrun bool) error {
	if w == nil {
		w = os.Stdout
	}
	const stale = 1 * time.Second
	report, err := dir.PurgeTemp(stale, dryrun)

	color := logs.Color(w)
	style := func(code, s string) string {
		if !color || s == "" {
			return s
		}
		return code + s + ansiReset
	}

	title := "Temporary Directory Purge"
	if dryrun {
		title += " (dry run)"
	}

	pluralDir := func(n int) string {
		if n == 1 {
			return "1 directory"
		}
		return fmt.Sprintf("%d directories", n)
	}

	pad := func(code, label string) string {
		return style(code, fmt.Sprintf("%-11s", label))
	}

	_, _ = fmt.Fprintln(w, style(ansiBold, title))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "  %s %s\n", pad(ansiCyan, "Pattern"), report.Pattern)
	_, _ = fmt.Fprintf(w, "  %s %s (%s)\n",
		pad(ansiCyan, "Found"),
		pluralDir(report.Found),
		helper.ByteCountFloat(report.FoundBytes))

	if report.Found > 0 {
		if dryrun {
			_, _ = fmt.Fprintf(w, "  %s %s (%s)\n",
				pad(ansiYellow, "Would purge"),
				pluralDir(report.Removed),
				helper.ByteCountFloat(report.RemovedBytes))
		} else {
			_, _ = fmt.Fprintf(w, "  %s %s (%s reclaimed)\n",
				pad(ansiGreen, "Removed"),
				pluralDir(report.Removed),
				helper.ByteCountFloat(report.RemovedBytes))
		}
	}

	if report.Skipped > 0 {
		_, _ = fmt.Fprintf(w, "  %s %s\n",
			pad(ansiDim, "Skipped"),
			style(ansiDim, fmt.Sprintf("%s (modified < %s ago)",
				pluralDir(report.Skipped), stale)))
	}
	if report.Failed > 0 {
		_, _ = fmt.Fprintf(w, "  %s %s\n",
			pad(ansiBoldRed, "Failed"),
			style(ansiBoldRed, pluralDir(report.Failed)))
	}

	if err != nil {
		return fmt.Errorf("purge temp dirs: %w", err)
	}
	return nil
}

// Issue describes a single configuration validation problem.
type Issue struct {
	Env   string
	Value string
	Issue string
}

// Issues returns all active validation problems across the configuration fields.
func (c Config) Issues() []Issue {
	var out []Issue
	add := func(env, val, issue string) {
		if issue != "" {
			out = append(out, Issue{Env: env, Value: val, Issue: issue})
		}
	}

	add("D2_DIR_LOG", c.AbsLog.String(), c.AbsLog.Issue())
	add("D2_DIR_DOWNLOAD", c.AbsDownload.String(), c.AbsDownload.Issue())
	add("D2_DIR_PREVIEW", c.AbsPreview.String(), c.AbsPreview.Issue())
	add("D2_DIR_THUMBNAIL", c.AbsThumbnail.String(), c.AbsThumbnail.Issue())
	add("D2_DIR_EXTRA", c.AbsExtra.String(), c.AbsExtra.Issue())
	add("D2_DIR_ORPHANED", c.AbsOrphaned.String(), c.AbsOrphaned.Issue())
	add("D2_TLS_CERT", c.TLSCert.String(), c.TLSCert.Issue())
	add("D2_TLS_KEY", c.TLSKey.String(), c.TLSKey.Issue())
	if c.TLSCert != "" && c.TLSKey == "" {
		add("D2_TLS_KEY", unsetLabel, "required when D2_TLS_CERT is set")
	}
	if c.TLSKey != "" && c.TLSCert == "" {
		add("D2_TLS_CERT", unsetLabel, "required when D2_TLS_KEY is set")
	}
	if issue := c.DemozooAuth.Issue(); issue != "" {
		add("D2_DEMOZOO_KEY", mask, issue)
	}
	return out
}

// Address returns the combined HTTP and HTTPS URLs the server may be reachable at.
func (c Config) Address() ([]string, error) {
	httpURLs, httpsURLs, err := c.addressURLs()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(httpURLs)+len(httpsURLs))
	out = append(out, httpURLs...)
	out = append(out, httpsURLs...)
	return out, nil
}
