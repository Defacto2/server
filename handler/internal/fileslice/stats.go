//nolint:gochecknoglobals
package fileslice

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
)

const StatsTTL = 10 * time.Minute // Stats time to live

var (
	statsCache     Stats
	statsCacheTime time.Time
	statsMu        sync.RWMutex
)

// Stats are the database statistics for the artifacts categories.
type Stats struct {
	Record    model.Artifacts
	Ansi      model.Ansi
	AnsiBBS   model.AnsiBBS
	BBS       model.BBS
	BBSText   model.BBSText
	BBStro    model.BBStro
	Console   model.Console
	Demoscene model.Demoscene
	MsDos     model.MsDos
	Intro     model.Intro
	IntroD    model.IntroMsDos
	IntroW    model.IntroWindows
	Installer model.Installer
	Java      model.Java
	Linux     model.Linux
	Magazine  model.Magazine
	Macos     model.Macos
	Nfo       model.Nfo
	NfoTool   model.NfoTool
	Proof     model.Proof
	Script    model.Script
	Text      model.Text
	Windows   model.Windows
}

const categoryStatsSQL = `SELECT
	COALESCE(BTRIM(section), ''),
	COALESCE(BTRIM(platform), ''),
	COUNT(*),
	COALESCE(SUM(filesize), 0),
	COALESCE(MIN(date_issued_year), 0),
	COALESCE(MAX(date_issued_year), 0)
FROM files
WHERE deletedat IS NULL
GROUP BY BTRIM(section), BTRIM(platform)`

func addStat(bytes, count, minYear, maxYear *int, rowBytes, rowCount, rowMin, rowMax int) {
	*bytes += rowBytes
	*count += rowCount
	if rowMin > 0 && (*minYear == 0 || rowMin < *minYear) {
		*minYear = rowMin
	}
	if rowMax > *maxYear {
		*maxYear = rowMax
	}
}

// Get and store the database statistics for the artifacts categories in a single scan.
func (s *Stats) Get(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const format = "category get stats %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	rows, err := exec.QueryContext(ctx, categoryStatsSQL)
	if err != nil {
		return fmt.Errorf(format, "query", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	uris := tags.URIs()
	var (
		sBBS     = uris[tags.BBS]
		sDemo    = uris[tags.Demo]
		sInstall = uris[tags.Install]
		sIntro   = uris[tags.Intro]
		sMag     = uris[tags.Mag]
		sNfo     = uris[tags.Nfo]
		sNfoTool = uris[tags.NfoTool]
		sProof   = uris[tags.Proof]
		pANSI    = uris[tags.ANSI]
		pConsole = uris[tags.Console]
		pDOS     = uris[tags.DOS]
		pJava    = uris[tags.Java]
		pLinux   = uris[tags.Linux]
		pMac     = uris[tags.Mac]
		pScript  = uris[tags.PHP]
		pText    = uris[tags.Text]
		pWindows = uris[tags.Windows]
	)

	for rows.Next() {
		var sect, plat string
		var cnt, sz, minY, maxY int
		if err := rows.Scan(&sect, &plat, &cnt, &sz, &minY, &maxY); err != nil {
			return fmt.Errorf(format, "scan", err)
		}

		addStat(&s.Record.Bytes, &s.Record.Count, &s.Record.MinYear, &s.Record.MaxYear, sz, cnt, minY, maxY)

		switch sect {
		case sBBS:
			addStat(&s.BBS.Bytes, &s.BBS.Count, &s.BBS.MinYear, &s.BBS.MaxYear, sz, cnt, minY, maxY)
			switch plat {
			case pANSI:
				addStat(&s.AnsiBBS.Bytes, &s.AnsiBBS.Count, &s.AnsiBBS.MinYear, &s.AnsiBBS.MaxYear, sz, cnt, minY, maxY)
			case pText:
				addStat(&s.BBSText.Bytes, &s.BBSText.Count, &s.BBSText.MinYear, &s.BBSText.MaxYear, sz, cnt, minY, maxY)
			case pDOS:
				addStat(&s.BBStro.Bytes, &s.BBStro.Count, &s.BBStro.MinYear, &s.BBStro.MaxYear, sz, cnt, minY, maxY)
			}
		case sDemo:
			addStat(&s.Demoscene.Bytes, &s.Demoscene.Count, &s.Demoscene.MinYear, &s.Demoscene.MaxYear, sz, cnt, minY, maxY)
		case sInstall:
			addStat(&s.Installer.Bytes, &s.Installer.Count, &s.Installer.MinYear, &s.Installer.MaxYear, sz, cnt, minY, maxY)
		case sIntro:
			addStat(&s.Intro.Bytes, &s.Intro.Count, &s.Intro.MinYear, &s.Intro.MaxYear, sz, cnt, minY, maxY)
			switch plat {
			case pDOS:
				addStat(&s.IntroD.Bytes, &s.IntroD.Count, &s.IntroD.MinYear, &s.IntroD.MaxYear, sz, cnt, minY, maxY)
			case pWindows:
				addStat(&s.IntroW.Bytes, &s.IntroW.Count, &s.IntroW.MinYear, &s.IntroW.MaxYear, sz, cnt, minY, maxY)
			}
		case sMag:
			addStat(&s.Magazine.Bytes, &s.Magazine.Count, &s.Magazine.MinYear, &s.Magazine.MaxYear, sz, cnt, minY, maxY)
		case sNfo:
			addStat(&s.Nfo.Bytes, &s.Nfo.Count, &s.Nfo.MinYear, &s.Nfo.MaxYear, sz, cnt, minY, maxY)
		case sNfoTool:
			addStat(&s.NfoTool.Bytes, &s.NfoTool.Count, &s.NfoTool.MinYear, &s.NfoTool.MaxYear, sz, cnt, minY, maxY)
		case sProof:
			addStat(&s.Proof.Bytes, &s.Proof.Count, &s.Proof.MinYear, &s.Proof.MaxYear, sz, cnt, minY, maxY)
		}

		switch plat {
		case pANSI:
			addStat(&s.Ansi.Bytes, &s.Ansi.Count, &s.Ansi.MinYear, &s.Ansi.MaxYear, sz, cnt, minY, maxY)
		case pConsole:
			addStat(&s.Console.Bytes, &s.Console.Count, &s.Console.MinYear, &s.Console.MaxYear, sz, cnt, minY, maxY)
		case pDOS:
			addStat(&s.MsDos.Bytes, &s.MsDos.Count, &s.MsDos.MinYear, &s.MsDos.MaxYear, sz, cnt, minY, maxY)
		case pJava:
			addStat(&s.Java.Bytes, &s.Java.Count, &s.Java.MinYear, &s.Java.MaxYear, sz, cnt, minY, maxY)
		case pLinux:
			addStat(&s.Linux.Bytes, &s.Linux.Count, &s.Linux.MinYear, &s.Linux.MaxYear, sz, cnt, minY, maxY)
		case pMac:
			addStat(&s.Macos.Bytes, &s.Macos.Count, &s.Macos.MinYear, &s.Macos.MaxYear, sz, cnt, minY, maxY)
		case pScript:
			addStat(&s.Script.Bytes, &s.Script.Count, &s.Script.MinYear, &s.Script.MaxYear, sz, cnt, minY, maxY)
		case pText:
			addStat(&s.Text.Bytes, &s.Text.Count, &s.Text.MinYear, &s.Text.MaxYear, sz, cnt, minY, maxY)
		case pWindows:
			addStat(&s.Windows.Bytes, &s.Windows.Count, &s.Windows.MinYear, &s.Windows.MaxYear, sz, cnt, minY, maxY)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf(format, "rows", err)
	}
	s.Record.UpdatedAt = time.Now()

	return nil
}

// Era is used to provide sorted statistics for artifact categories.
type Era struct {
	Name    string
	MinYear int // MinYear is the earliest or founding year
	MaxYear int // MaxYear is the most recent or final year
}

// SortYear returns the Eras sorted from oldest to newest.
// If multiple Era share the same year, then the shortest to longest time span is used.
//
// For example:
//   - 1. 1980 - 1982
//   - 2. 1980 - 1990
//   - 3. 1981 - 1985
func (s *Stats) SortYear() [22]Era {
	items := s.eras()

	slices.SortFunc(items[:], func(a, b Era) int {
		if a.MinYear != b.MinYear {
			return cmp.Compare(a.MinYear, b.MinYear)
		}
		// If min years are equal, order by recent year
		return cmp.Compare(a.MaxYear, b.MaxYear)
	})

	return items
}

// Item is used to provide sorted statistics for artifact categories.
type Item struct {
	Name    string
	URI     string
	Bytes   int
	Count   int
	MinYear int
	MaxYear int
}

// SortByte returns the Items sorted from highest to lowest bytes.
func (s *Stats) SortByte() [22]Item {
	items := s.items()

	slices.SortFunc(items[:], func(a, b Item) int {
		if a.Bytes != b.Bytes {
			return cmp.Compare(b.Bytes, a.Bytes)
		}
		// If bytes are equal, order by count
		return cmp.Compare(b.Count, a.Count)
	})

	return items
}

// SortCount returns the Items sorted by highest to lowest counts.
func (s *Stats) SortCount() [22]Item {
	items := s.items()

	slices.SortFunc(items[:], func(a, b Item) int {
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		// If count are equal, order by desc byte size
		return cmp.Compare(b.Bytes, a.Bytes)
	})

	return items
}

// SortName returns the Items sorted alphabetically by their name.
func (s *Stats) SortName() [22]Item {
	items := s.items()

	slices.SortFunc(items[:], func(a, b Item) int {
		return strings.Compare(a.Name, b.Name)
	})

	return items
}

func (s *Stats) eras() [22]Era {
	return [...]Era{
		{s.Ansi.String(), s.Ansi.MinYear, s.Ansi.MaxYear},
		{s.AnsiBBS.String(), s.AnsiBBS.MinYear, s.AnsiBBS.MaxYear},
		{s.BBS.String(), s.BBS.MinYear, s.BBS.MaxYear},
		{s.BBSText.String(), s.BBSText.MinYear, s.BBSText.MaxYear},
		{s.BBStro.String(), s.BBStro.MinYear, s.BBStro.MaxYear},
		{s.Console.String(), s.Console.MinYear, s.Console.MaxYear},
		{s.Demoscene.String(), s.Demoscene.MinYear, s.Demoscene.MaxYear},
		{s.MsDos.String(), s.MsDos.MinYear, s.MsDos.MaxYear},
		{s.Intro.String(), s.Intro.MinYear, s.Intro.MaxYear},
		{s.IntroD.String(), s.IntroD.MinYear, s.IntroD.MaxYear},
		{s.IntroW.String(), s.IntroW.MinYear, s.IntroW.MaxYear},
		{s.Installer.String(), s.Installer.MinYear, s.Installer.MaxYear},
		{s.Java.String(), s.Java.MinYear, s.Java.MaxYear},
		{s.Linux.String(), s.Linux.MinYear, s.Linux.MaxYear},
		{s.Magazine.String(), s.Magazine.MinYear, s.Magazine.MaxYear},
		{s.Macos.String(), s.Macos.MinYear, s.Macos.MaxYear},
		{s.Nfo.String(), s.Nfo.MinYear, s.Nfo.MaxYear},
		{s.NfoTool.String(), s.NfoTool.MinYear, s.NfoTool.MaxYear},
		{s.Proof.String(), s.Proof.MinYear, s.Proof.MaxYear},
		{s.Script.String(), s.Script.MinYear, s.Script.MaxYear},
		{s.Text.String(), s.Text.MinYear, s.Text.MaxYear},
		{s.Windows.String(), s.Windows.MinYear, s.Windows.MaxYear},
	}
}

func (s *Stats) items() [22]Item {
	return [...]Item{
		{s.Ansi.String(), "ansi", s.Ansi.Bytes, s.Ansi.Count, s.Ansi.MinYear, s.Ansi.MaxYear},
		{s.AnsiBBS.String(), "ansi-bbs", s.AnsiBBS.Bytes, s.AnsiBBS.Count, s.AnsiBBS.MinYear, s.AnsiBBS.MaxYear},
		{s.BBS.String(), "bbs", s.BBS.Bytes, s.BBS.Count, s.BBS.MinYear, s.BBS.MaxYear},
		{s.BBSText.String(), "bbs-text", s.BBSText.Bytes, s.BBSText.Count, s.BBSText.MinYear, s.BBSText.MaxYear},
		{s.BBStro.String(), "bbstro", s.BBStro.Bytes, s.BBStro.Count, s.BBStro.MinYear, s.BBStro.MaxYear},
		{s.Console.String(), "console", s.Console.Bytes, s.Console.Count, s.Console.MinYear, s.Console.MaxYear},
		{s.Demoscene.String(), "demoscene", s.Demoscene.Bytes, s.Demoscene.Count, s.Demoscene.MinYear, s.Demoscene.MaxYear},
		{s.MsDos.String(), "msdos", s.MsDos.Bytes, s.MsDos.Count, s.MsDos.MinYear, s.MsDos.MaxYear},
		{s.Intro.String(), "intro", s.Intro.Bytes, s.Intro.Count, s.Intro.MinYear, s.Intro.MaxYear},
		{s.IntroD.String(), "intro-msdos", s.IntroD.Bytes, s.IntroD.Count, s.IntroD.MinYear, s.IntroD.MaxYear},
		{s.IntroW.String(), "intro-windows", s.IntroW.Bytes, s.IntroW.Count, s.IntroW.MinYear, s.IntroW.MaxYear},
		{s.Installer.String(), "installer", s.Installer.Bytes, s.Installer.Count, s.Installer.MinYear, s.Installer.MaxYear},
		{s.Java.String(), "java", s.Java.Bytes, s.Java.Count, s.Java.MinYear, s.Java.MaxYear},
		{s.Linux.String(), "linux", s.Linux.Bytes, s.Linux.Count, s.Linux.MinYear, s.Linux.MaxYear},
		{s.Magazine.String(), "magazine", s.Magazine.Bytes, s.Magazine.Count, s.Magazine.MinYear, s.Magazine.MaxYear},
		{s.Macos.String(), "macos", s.Macos.Bytes, s.Macos.Count, s.Macos.MinYear, s.Macos.MaxYear},
		{s.Nfo.String(), "nfo", s.Nfo.Bytes, s.Nfo.Count, s.Nfo.MinYear, s.Nfo.MaxYear},
		{s.NfoTool.String(), "nfo-tool", s.NfoTool.Bytes, s.NfoTool.Count, s.NfoTool.MinYear, s.NfoTool.MaxYear},
		{s.Proof.String(), "proof", s.Proof.Bytes, s.Proof.Count, s.Proof.MinYear, s.Proof.MaxYear},
		{s.Script.String(), "script", s.Script.Bytes, s.Script.Count, s.Script.MinYear, s.Script.MaxYear},
		{s.Text.String(), "text", s.Text.Bytes, s.Text.Count, s.Text.MinYear, s.Text.MaxYear},
		{s.Windows.String(), "windows", s.Windows.Bytes, s.Windows.Count, s.Windows.MinYear, s.Windows.MaxYear},
	}
}

// Counter returns the statistics for the artifacts categories, cached for 10 minutes.
func Counter(ctx context.Context, sl *slog.Logger, db *sql.DB) (Stats, error) {
	const format = "artifacts categories counter %s: %w"

	if err := nils.Check(ctx, sl, db); err != nil {
		return Stats{}, fmt.Errorf(format, "check", err)
	}

	statsMu.RLock()
	if time.Since(statsCacheTime) < StatsTTL {
		cached := statsCache
		statsMu.RUnlock()
		return cached, nil
	}
	statsMu.RUnlock()

	statsMu.Lock()
	defer statsMu.Unlock()

	if time.Since(statsCacheTime) < StatsTTL {
		return statsCache, nil
	}

	counter := Statistics()
	if err := counter.Get(ctx, sl, db); err != nil {
		return Stats{}, fmt.Errorf(format, "get", err)
	}

	statsCache = counter
	statsCacheTime = time.Now()

	return statsCache, nil
}

// Statistics returns the empty database statistics for the artifacts categories.
func Statistics() Stats {
	return Stats{} //nolint:exhaustruct_v5
}
