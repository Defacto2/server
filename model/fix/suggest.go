package fix

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
)

const (
	suggestCacheTTL  = 5 * time.Minute
	minNormPrefixLen = 3
	minTypoLen       = 6
)

// Suggestion is a pre-normalized releaser cached in memory for fast auto-suggestions.
type Suggestion struct {
	name        string
	count       int
	fingerprint string
	sortedWords string
	core        string
}

type suggestCache struct {
	mu          sync.RWMutex
	all         []Suggestion
	allLoadedAt time.Time
	mag         []Suggestion
	magLoadedAt time.Time
}

//nolint:gochecknoglobals
var relSug suggestCache

// Clear clears the in-memory releaser suggestion cache so the next
// lookup refreshes from the database.
func Clear() {
	relSug.mu.Lock()
	relSug.all = nil
	relSug.allLoadedAt = time.Time{}
	relSug.mag = nil
	relSug.magLoadedAt = time.Time{}
	relSug.mu.Unlock()
}

// SuggestReleasers returns up to limit releaser suggestions matching lookups,
// ranked by raw prefix match, normalized duplicate-aware match, and 1-char typo match,
// ordered by file count within each tier.
func SuggestReleasers(
	ctx context.Context, exec boil.ContextExecutor, magazine bool, limit int, lookups ...string,
) ([]ReleaserCount, error) {
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf("suggest releasers: %w", err)
	}
	if len(lookups) == 0 || limit <= 0 {
		return nil, nil
	}

	entries, err := suggestions(ctx, exec, magazine)
	if err != nil {
		return nil, fmt.Errorf("suggest releasers load: %w", err)
	}

	return Suggests(entries, limit, lookups...), nil
}

func suggestions(
	ctx context.Context, exec boil.ContextExecutor, magazine bool,
) ([]Suggestion, error) {
	now := time.Now()

	relSug.mu.RLock()
	if magazine && relSug.mag != nil && now.Sub(relSug.magLoadedAt) < suggestCacheTTL {
		entries := relSug.mag
		relSug.mu.RUnlock()
		return entries, nil
	}
	if !magazine && relSug.all != nil && now.Sub(relSug.allLoadedAt) < suggestCacheTTL {
		entries := relSug.all
		relSug.mu.RUnlock()
		return entries, nil
	}
	relSug.mu.RUnlock()

	query := string(postgres.SuggestReleasersCounts())
	if magazine {
		query = string(postgres.SuggestMagazinesCounts())
	}

	var rows []ReleaserCount
	if err := queries.Raw(query).Bind(ctx, exec, &rows); err != nil {
		return nil, err
	}

	entries := Suggest(rows)

	relSug.mu.Lock()
	if magazine {
		relSug.mag = entries
		relSug.magLoadedAt = now
	} else {
		relSug.all = entries
		relSug.allLoadedAt = now
	}
	relSug.mu.Unlock()

	return entries, nil
}

// Suggest pre-normalizes releaser rows and sorts them by file count descending.
func Suggest(rows []ReleaserCount) []Suggestion {
	entries := make([]Suggestion, 0, len(rows))
	for _, r := range rows {
		raw := strings.TrimSpace(r.Releaser)
		if raw == "" {
			continue
		}
		kind := classifyKind(raw)
		fp, sortedW, core := normalizeTokens(raw, kind)
		entries = append(entries, Suggestion{
			name:        strings.ToUpper(raw),
			count:       r.Count,
			fingerprint: fp,
			sortedWords: sortedW,
			core:        core,
		})
	}
	slices.SortFunc(entries, func(a, b Suggestion) int {
		if c := cmp.Compare(b.count, a.count); c != 0 {
			return c
		}
		return cmp.Compare(a.name, b.name)
	})
	return entries
}

// Suggests matches lookups against pre-normalized entries in 3 priority tiers:
//  1. Raw prefix match (exact existing behavior)
//  2. Normalized fingerprint match (handles stop words, punctuation, apostrophes, numerals)
//  3. 1-character typo or adjacent letter transposition (for normalized inputs >= 6 chars)
func Suggests(entries []Suggestion, limit int, lookups ...string) []ReleaserCount {
	if len(entries) == 0 || len(lookups) == 0 || limit <= 0 {
		return nil
	}

	rawPrefixes := make([]string, 0, len(lookups)*3)
	normPrefixes := make([]string, 0, len(lookups))
	normSorted := make([]string, 0, len(lookups))
	for _, name := range lookups {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		rawPrefixes = append(rawPrefixes,
			strings.ToUpper(trimmed),
			strings.ToUpper(releaser.Title(trimmed)),
			strings.ToUpper(releaser.Cell(trimmed)),
		)
		kind := classifyKind(trimmed)
		fp, sortedW, _ := normalizeTokens(trimmed, kind)
		if len(fp) >= minNormPrefixLen {
			normPrefixes = append(normPrefixes, fp)
		}
		if sortedW != "" {
			normSorted = append(normSorted, sortedW)
		}
	}
	if len(rawPrefixes) == 0 {
		return nil
	}
	slices.Sort(rawPrefixes)
	rawPrefixes = slices.Compact(rawPrefixes)
	slices.Sort(normPrefixes)
	normPrefixes = slices.Compact(normPrefixes)
	slices.Sort(normSorted)
	normSorted = slices.Compact(normSorted)

	tier1 := make([]ReleaserCount, 0, limit)
	tier2 := make([]ReleaserCount, 0, limit)
	tier3 := make([]ReleaserCount, 0, limit)

	for _, item := range entries {
		if hasAnyPrefix(item.name, rawPrefixes) {
			tier1 = append(tier1, ReleaserCount{Releaser: item.name, Count: item.count})
			if len(tier1) >= limit {
				return tier1
			}
			continue
		}

		if len(normPrefixes) == 0 || len(item.fingerprint) < minNormPrefixLen {
			continue
		}

		if len(tier1)+len(tier2) < limit &&
			(hasAnyPrefix(item.fingerprint, normPrefixes) ||
				(item.sortedWords != "" && slices.Contains(normSorted, item.sortedWords))) {
			tier2 = append(tier2, ReleaserCount{Releaser: item.name, Count: item.count})
			continue
		}

		if len(tier1)+len(tier2)+len(tier3) < limit && len(item.core) >= minTypoLen {
			if matchesTypo(item.core, normPrefixes) {
				tier3 = append(tier3, ReleaserCount{Releaser: item.name, Count: item.count})
			}
		}
	}

	out := make([]ReleaserCount, 0, min(limit, len(tier1)+len(tier2)+len(tier3)))
	out = append(out, tier1...)
	if rem := limit - len(out); rem > 0 && len(tier2) > 0 {
		out = append(out, tier2[:min(rem, len(tier2))]...)
	}
	if rem := limit - len(out); rem > 0 && len(tier3) > 0 {
		out = append(out, tier3[:min(rem, len(tier3))]...)
	}
	return out
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func matchesTypo(core string, normPrefixes []string) bool {
	for _, fp := range normPrefixes {
		if len(fp) < minTypoLen {
			continue
		}
		diff := len(core) - len(fp)
		if diff < -1 || diff > 1 {
			continue
		}
		if singleTypo(core, fp) {
			return true
		}
	}
	return false
}
