//nolint:gochecknoglobals
package fileslice

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"golang.org/x/sync/errgroup"
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

// Get and store the database statistics for the artifacts categories.
func (s *Stats) Get(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "category get stats %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	const limit = 10 // max database connections
	g.SetLimit(limit)

	// fetch record
	g.Go(func() error {
		if err := s.Record.Public(ctx, exec); err != nil {
			return fmt.Errorf(format, "record", err)
		}
		return nil
	})

	// concurrent, individual category stats
	stats := []struct {
		name string
		fn   func(context.Context, boil.ContextExecutor) error
	}{
		{"ansi", s.Ansi.Stat},
		{"ansi bbs", s.AnsiBBS.Stat},
		{"bbs", s.BBS.Stat},
		{"bbs text", s.BBSText.Stat},
		{"bbstro", s.BBStro.Stat},
		{"console", s.Console.Stat},
		{"ms-dos", s.MsDos.Stat},
		{"intro", s.Intro.Stat},
		{"intro ms-dos", s.IntroD.Stat},
		{"intro windows", s.IntroW.Stat},
		{"installer", s.Installer.Stat},
		{"java", s.Java.Stat},
		{"linux", s.Linux.Stat},
		{"demoscene", s.Demoscene.Stat},
		{"macos", s.Macos.Stat},
		{"magazine", s.Magazine.Stat},
		{"nfo", s.Nfo.Stat},
		{"nfo tool", s.NfoTool.Stat},
		{"proof", s.Proof.Stat},
		{"script", s.Script.Stat},
		{"text", s.Text.Stat},
		{"windows", s.Windows.Stat},
	}

	for _, st := range stats {
		g.Go(func() error {
			if err := st.fn(ctx, exec); err != nil {
				return fmt.Errorf(format, st.name, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf(format, "wait", err)
	}

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
func (s *Stats) SortYear() [20]Era {
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

// Item is used to provided sorted statistics for artifact categories.
type Item struct {
	Name  string
	Bytes int
	Count int
}

// SortByte returns the Items sorted from highest to lowest bytes.
func (s *Stats) SortByte() [21]Item {
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
func (s *Stats) SortCount() [21]Item {
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
func (s *Stats) SortName() [21]Item {
	items := s.items()

	slices.SortFunc(items[:], func(a, b Item) int {
		return strings.Compare(a.Name, b.Name)
	})

	return items
}

func (s *Stats) eras() [20]Era {
	return [...]Era{
		{s.Ansi.String(), s.Ansi.MinYear, s.Ansi.MaxYear},
		{s.AnsiBBS.String(), s.AnsiBBS.MinYear, s.AnsiBBS.MaxYear},
		{s.BBS.String(), s.BBS.MinYear, s.BBS.MaxYear},
		{s.BBSText.String(), s.BBSText.MinYear, s.BBSText.MaxYear},
		{s.BBStro.String(), s.BBStro.MinYear, s.BBStro.MaxYear},
		//	{s.Console.String(), s.Console.MinYear, s.Console.MaxYear},
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
		{s.Windows.String(), s.Windows.MinYear, s.Windows.MaxYear},
	}
}

func (s *Stats) items() [21]Item {
	return [...]Item{
		{s.Ansi.String(), s.Ansi.Bytes, s.Ansi.Count},
		{s.AnsiBBS.String(), s.AnsiBBS.Bytes, s.AnsiBBS.Count},
		{s.BBS.String(), s.BBS.Bytes, s.BBS.Count},
		{s.BBSText.String(), s.BBSText.Bytes, s.BBSText.Count},
		{s.BBStro.String(), s.BBStro.Bytes, s.BBStro.Count},
		{s.Console.String(), s.Console.Bytes, s.Console.Count},
		{s.Demoscene.String(), s.Demoscene.Bytes, s.Demoscene.Count},
		{s.MsDos.String(), s.MsDos.Bytes, s.MsDos.Count},
		{s.Intro.String(), s.Intro.Bytes, s.Intro.Count},
		{s.IntroD.String(), s.IntroD.Bytes, s.IntroD.Count},
		{s.IntroW.String(), s.IntroW.Bytes, s.IntroW.Count},
		{s.Installer.String(), s.Installer.Bytes, s.Installer.Count},
		{s.Java.String(), s.Java.Bytes, s.Java.Count},
		{s.Linux.String(), s.Linux.Bytes, s.Linux.Count},
		{s.Magazine.String(), s.Magazine.Bytes, s.Magazine.Count},
		{s.Macos.String(), s.Macos.Bytes, s.Macos.Count},
		{s.Nfo.String(), s.Nfo.Bytes, s.Nfo.Count},
		{s.NfoTool.String(), s.NfoTool.Bytes, s.NfoTool.Count},
		{s.Proof.String(), s.Proof.Bytes, s.Proof.Count},
		{s.Script.String(), s.Script.Bytes, s.Script.Count},
		{s.Windows.String(), s.Windows.Bytes, s.Windows.Count},
	}
}

// Counter returns the statistics for the artifacts categories, cached for 10 minutes.
func Counter(ctx context.Context, db *sql.DB) (Stats, error) {
	const format = "artifacts categories counter %s: %w"

	if err := nils.Check(ctx, db); err != nil {
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
	if err := counter.Get(ctx, db); err != nil {
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
