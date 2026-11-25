//nolint:funcorder,gochecknoglobals
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

const (
	StatsTTL  = 10 * time.Minute // Stats time to live
	YearlyMin = 1980             // YearlyMin is the first year included in the yearly release series.
	YearlyMax = 2010             // YearlyMax is the last year included in the yearly release series.
)

// YearlyPoint holds aggregate release counts and bytes for a single calendar year.
type YearlyPoint struct {
	Year    int   `json:"year"`
	Total   int   `json:"total"`
	Bytes   int64 `json:"bytes"`
	DOS     int   `json:"dos"`
	Windows int   `json:"windows"`
	Text    int   `json:"text"`
}

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
	Yearly    []YearlyPoint
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

const yearlyStatsSQL = `SELECT
	date_issued_year,
	COALESCE(BTRIM(platform), ''),
	COUNT(*),
	COALESCE(SUM(filesize), 0)
FROM files
WHERE deletedat IS NULL
  AND date_issued_year BETWEEN 1980 AND 2010
GROUP BY date_issued_year, BTRIM(platform)
ORDER BY date_issued_year ASC`

type statRow struct {
	sect, plat          string
	cnt, sz, minY, maxY int
}

type statBucket struct {
	bytes, count, minYear, maxYear *int
}

func newBucket(bytes, count, minYear, maxYear *int) statBucket {
	return statBucket{
		bytes:   bytes,
		count:   count,
		minYear: minYear,
		maxYear: maxYear,
	}
}

func (b statBucket) add(r statRow) {
	*b.bytes += r.sz
	*b.count += r.cnt
	if r.minY > 0 && (*b.minYear == 0 || r.minY < *b.minYear) {
		*b.minYear = r.minY
	}
	if r.maxY > *b.maxYear {
		*b.maxYear = r.maxY
	}
}

func (s *Stats) sectionBuckets(uris map[tags.Tag]string) map[string]statBucket {
	return map[string]statBucket{
		uris[tags.BBS]:     newBucket(&s.BBS.Bytes, &s.BBS.Count, &s.BBS.MinYear, &s.BBS.MaxYear),
		uris[tags.Demo]:    newBucket(&s.Demoscene.Bytes, &s.Demoscene.Count, &s.Demoscene.MinYear, &s.Demoscene.MaxYear),
		uris[tags.Install]: newBucket(&s.Installer.Bytes, &s.Installer.Count, &s.Installer.MinYear, &s.Installer.MaxYear),
		uris[tags.Intro]:   newBucket(&s.Intro.Bytes, &s.Intro.Count, &s.Intro.MinYear, &s.Intro.MaxYear),
		uris[tags.Mag]:     newBucket(&s.Magazine.Bytes, &s.Magazine.Count, &s.Magazine.MinYear, &s.Magazine.MaxYear),
		uris[tags.Nfo]:     newBucket(&s.Nfo.Bytes, &s.Nfo.Count, &s.Nfo.MinYear, &s.Nfo.MaxYear),
		uris[tags.NfoTool]: newBucket(&s.NfoTool.Bytes, &s.NfoTool.Count, &s.NfoTool.MinYear, &s.NfoTool.MaxYear),
		uris[tags.Proof]:   newBucket(&s.Proof.Bytes, &s.Proof.Count, &s.Proof.MinYear, &s.Proof.MaxYear),
	}
}

func (s *Stats) platformBuckets(uris map[tags.Tag]string) map[string]statBucket {
	return map[string]statBucket{
		uris[tags.ANSI]:    newBucket(&s.Ansi.Bytes, &s.Ansi.Count, &s.Ansi.MinYear, &s.Ansi.MaxYear),
		uris[tags.Console]: newBucket(&s.Console.Bytes, &s.Console.Count, &s.Console.MinYear, &s.Console.MaxYear),
		uris[tags.DOS]:     newBucket(&s.MsDos.Bytes, &s.MsDos.Count, &s.MsDos.MinYear, &s.MsDos.MaxYear),
		uris[tags.Java]:    newBucket(&s.Java.Bytes, &s.Java.Count, &s.Java.MinYear, &s.Java.MaxYear),
		uris[tags.Linux]:   newBucket(&s.Linux.Bytes, &s.Linux.Count, &s.Linux.MinYear, &s.Linux.MaxYear),
		uris[tags.Mac]:     newBucket(&s.Macos.Bytes, &s.Macos.Count, &s.Macos.MinYear, &s.Macos.MaxYear),
		uris[tags.PHP]:     newBucket(&s.Script.Bytes, &s.Script.Count, &s.Script.MinYear, &s.Script.MaxYear),
		uris[tags.Text]:    newBucket(&s.Text.Bytes, &s.Text.Count, &s.Text.MinYear, &s.Text.MaxYear),
		uris[tags.Windows]: newBucket(&s.Windows.Bytes, &s.Windows.Count, &s.Windows.MinYear, &s.Windows.MaxYear),
	}
}

func (s *Stats) pairBuckets(uris map[tags.Tag]string) map[[2]string]statBucket {
	return map[[2]string]statBucket{
		{uris[tags.BBS], uris[tags.ANSI]}:      newBucket(&s.AnsiBBS.Bytes, &s.AnsiBBS.Count, &s.AnsiBBS.MinYear, &s.AnsiBBS.MaxYear),
		{uris[tags.BBS], uris[tags.Text]}:      newBucket(&s.BBSText.Bytes, &s.BBSText.Count, &s.BBSText.MinYear, &s.BBSText.MaxYear),
		{uris[tags.BBS], uris[tags.DOS]}:       newBucket(&s.BBStro.Bytes, &s.BBStro.Count, &s.BBStro.MinYear, &s.BBStro.MaxYear),
		{uris[tags.Intro], uris[tags.DOS]}:     newBucket(&s.IntroD.Bytes, &s.IntroD.Count, &s.IntroD.MinYear, &s.IntroD.MaxYear),
		{uris[tags.Intro], uris[tags.Windows]}: newBucket(&s.IntroW.Bytes, &s.IntroW.Count, &s.IntroW.MinYear, &s.IntroW.MaxYear),
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
	record := newBucket(&s.Record.Bytes, &s.Record.Count, &s.Record.MinYear, &s.Record.MaxYear)
	bySection := s.sectionBuckets(uris)
	byPlatform := s.platformBuckets(uris)
	byPair := s.pairBuckets(uris)

	for rows.Next() {
		var r statRow
		if err := rows.Scan(&r.sect, &r.plat, &r.cnt, &r.sz, &r.minY, &r.maxY); err != nil {
			return fmt.Errorf(format, "scan", err)
		}
		record.add(r)
		if b, ok := bySection[r.sect]; ok {
			b.add(r)
		}
		if b, ok := byPlatform[r.plat]; ok {
			b.add(r)
		}
		if b, ok := byPair[[2]string{r.sect, r.plat}]; ok {
			b.add(r)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf(format, "rows", err)
	}

	if err := s.getYearly(ctx, exec, uris[tags.DOS], uris[tags.Windows], uris[tags.ANSI], uris[tags.Text]); err != nil {
		return fmt.Errorf(format, "yearly", err)
	}
	s.Record.UpdatedAt = time.Now()

	return nil
}

func (s *Stats) getYearly(ctx context.Context, exec boil.ContextExecutor, pDOS, pWindows, pANSI, pText string) error {
	rows, err := exec.QueryContext(ctx, yearlyStatsSQL)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	points := make([]YearlyPoint, YearlyMax-YearlyMin+1)
	for i := range points {
		points[i].Year = YearlyMin + i
	}

	for rows.Next() {
		var yr, cnt int
		var plat string
		var sz int64
		if err := rows.Scan(&yr, &plat, &cnt, &sz); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		if yr < YearlyMin || yr > YearlyMax {
			continue
		}
		pt := &points[yr-YearlyMin]
		pt.Total += cnt
		pt.Bytes += sz
		switch plat {
		case pDOS:
			pt.DOS += cnt
		case pWindows:
			pt.Windows += cnt
		case pANSI, pText:
			pt.Text += cnt
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}

	s.Yearly = points
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
