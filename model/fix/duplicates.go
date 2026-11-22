package fix

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
)

// ReleaserCount holds a raw releaser name and its file count from the database.
type ReleaserCount struct {
	Releaser string `boil:"releaser"`
	Count    int    `boil:"count_sum"`
}

// DupEntry represents a single releaser candidate within a duplicate cluster.
type DupEntry struct {
	Name    string
	Display string
	URI     string
	Kind    string
	Count   int
}

// DupGroup represents a cluster of two or more potentially duplicate releaser names.
type DupGroup struct {
	Key     string
	Kind    string
	Tier    int
	Reason  string
	Entries []DupEntry
}

// DupReport summarizes all detected duplicate clusters across confidence tiers.
type DupReport struct {
	TotalReleasers int
	TotalClusters  int
	Tier1          []DupGroup
	Tier2          []DupGroup
	Tier3          []DupGroup
}

// Duplicates queries all distinct public releasers and clusters potential duplicates.
func Duplicates(ctx context.Context, exec boil.ContextExecutor) (*DupReport, error) {
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf("fix duplicates: %w", err)
	}
	var rows []ReleaserCount
	if err := queries.Raw(string(postgres.AllReleasersCounts())).Bind(ctx, exec, &rows); err != nil {
		return nil, fmt.Errorf("fix duplicates query: %w", err)
	}
	return AnalyzeDuplicates(rows), nil
}

// AnalyzeDuplicates runs the 3-tier in-memory clustering pipeline on a slice of releaser counts.
func AnalyzeDuplicates(rows []ReleaserCount) *DupReport {
	report := &DupReport{
		TotalReleasers: len(rows),
	}
	if len(rows) < 2 {
		return report
	}

	type parsedItem struct {
		entry       DupEntry
		fingerprint string
		sortedWords string
		core        string
	}

	items := make([]parsedItem, 0, len(rows))
	for _, r := range rows {
		raw := strings.TrimSpace(r.Releaser)
		if raw == "" {
			continue
		}
		rawLabel := r.Releaser
		if r.Releaser != raw {
			rawLabel = fmt.Sprintf("%q (untrimmed whitespace)", r.Releaser)
		}
		kind := classifyKind(raw)
		uri := helper.Slug(raw)
		disp := releaser.Index(raw)
		if disp == "" {
			disp = helper.Capitalize(strings.ToLower(raw))
		}
		fp, sortedW, core := normalizeTokens(raw, kind)
		items = append(items, parsedItem{
			entry: DupEntry{
				Name:    rawLabel,
				Display: disp,
				URI:     uri,
				Kind:    kind,
				Count:   r.Count,
			},
			fingerprint: fp,
			sortedWords: sortedW,
			core:        core,
		})
	}

	// Track pairs already grouped in a higher tier so we don't repeat them.
	seenPairs := make(map[string]struct{})
	markCluster := func(entries []DupEntry) {
		for i := range entries {
			for j := i + 1; j < len(entries); j++ {
				seenPairs[pairKey(entries[i].Name, entries[j].Name)] = struct{}{}
			}
		}
	}
	allSeen := func(entries []DupEntry) bool {
		for i := range entries {
			for j := i + 1; j < len(entries); j++ {
				if _, ok := seenPairs[pairKey(entries[i].Name, entries[j].Name)]; !ok {
					return false
				}
			}
		}
		return true
	}

	// Tier 1: Normalized fingerprint collisions.
	tier1Buckets := make(map[string][]DupEntry)
	tier1Order := make([]string, 0)
	for _, it := range items {
		if len(it.fingerprint) < 3 {
			continue
		}
		k := it.entry.Kind + ":" + it.fingerprint
		if _, exists := tier1Buckets[k]; !exists {
			tier1Order = append(tier1Order, k)
		}
		tier1Buckets[k] = append(tier1Buckets[k], it.entry)
	}
	for _, k := range tier1Order {
		group := tier1Buckets[k]
		if len(group) < 2 {
			continue
		}
		sortEntries(group)
		markCluster(group)
		report.Tier1 = append(report.Tier1, DupGroup{
			Key:     strings.TrimPrefix(k, group[0].Kind+":"),
			Kind:    group[0].Kind,
			Tier:    1,
			Reason:  "Normalized match",
			Entries: group,
		})
	}

	// Tier 2: Token order / word swaps.
	tier2Buckets := make(map[string][]DupEntry)
	tier2Order := make([]string, 0)
	for _, it := range items {
		if it.sortedWords == "" || !strings.Contains(it.sortedWords, " ") {
			continue
		}
		k := it.entry.Kind + ":" + it.sortedWords
		if _, exists := tier2Buckets[k]; !exists {
			tier2Order = append(tier2Order, k)
		}
		tier2Buckets[k] = append(tier2Buckets[k], it.entry)
	}
	for _, k := range tier2Order {
		group := tier2Buckets[k]
		if len(group) < 2 || allSeen(group) {
			continue
		}
		sortEntries(group)
		markCluster(group)
		report.Tier2 = append(report.Tier2, DupGroup{
			Key:     strings.TrimPrefix(k, group[0].Kind+":"),
			Kind:    group[0].Kind,
			Tier:    2,
			Reason:  "Word / token swap",
			Entries: group,
		})
	}

	// Tier 3: 1-character typo or adjacent letter transposition within (Kind, 2-char prefix) buckets.
	const minCoreLen = 6
	prefixBuckets := make(map[string][]parsedItem)
	prefixOrder := make([]string, 0)
	for _, it := range items {
		if len(it.core) < minCoreLen {
			continue
		}
		pk := it.entry.Kind + ":" + it.core[:2]
		if _, exists := prefixBuckets[pk]; !exists {
			prefixOrder = append(prefixOrder, pk)
		}
		prefixBuckets[pk] = append(prefixBuckets[pk], it)
	}

	for _, pk := range prefixOrder {
		bucket := prefixBuckets[pk]
		if len(bucket) < 2 {
			continue
		}
		for i := range bucket {
			a := bucket[i]
			for j := i + 1; j < len(bucket); j++ {
				b := bucket[j]
				if a.core == b.core {
					continue
				}
				lenDiff := len(a.core) - len(b.core)
				if lenDiff < -1 || lenDiff > 1 {
					continue
				}
				if _, ok := seenPairs[pairKey(a.entry.Name, b.entry.Name)]; ok {
					continue
				}
				if isSingleTypo(a.core, b.core) {
					pair := []DupEntry{a.entry, b.entry}
					sortEntries(pair)
					markCluster(pair)
					report.Tier3 = append(report.Tier3, DupGroup{
						Key:     a.core + " ~ " + b.core,
						Kind:    a.entry.Kind,
						Tier:    3,
						Reason:  "1-char typo / swap",
						Entries: pair,
					})
				}
			}
		}
	}

	report.TotalClusters = len(report.Tier1) + len(report.Tier2) + len(report.Tier3)
	return report
}

func classifyKind(name string) string {
	u := strings.ToUpper(strings.TrimSpace(name))
	switch {
	case strings.HasSuffix(u, " BBS") || u == "BBS":
		return "BBS"
	case strings.HasSuffix(u, " FTP") || u == "FTP":
		return "FTP"
	default:
		return "Group"
	}
}

func pairKey(a, b string) string {
	if a < b {
		return a + "\x00" + b
	}
	return b + "\x00" + a
}

func sortEntries(entries []DupEntry) {
	slices.SortFunc(entries, func(a, b DupEntry) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Name, b.Name)
	})
}

func isStopWord(w string) bool {
	switch w {
	case "the", "a", "an", "and", "n", "of",
		"inc", "incorporated", "ltd", "limited", "corp", "corporation",
		"co", "company", "llc", "plc",
		"group", "crew", "team", "club",
		"productions", "production", "prod", "prods",
		"software", "soft", "designs", "design",
		"division", "div", "international", "intl",
		"organization", "org", "enterprise", "enterprises",
		"association", "systems", "system", "sys":
		return true
	default:
		return false
	}
}

func normalizeNumeral(w string) string {
	switch w {
	case "ii", "two":
		return "2"
	case "iii", "three":
		return "3"
	case "iv", "four":
		return "4"
	case "vi", "six":
		return "6"
	case "vii", "seven":
		return "7"
	case "viii", "eight":
		return "8"
	case "ix", "nine":
		return "9"
	case "one":
		return "1"
	default:
		return w
	}
}

func normalizeTokens(raw, kind string) (fingerprint, sortedWords, core string) {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch kind {
	case "BBS":
		s = strings.TrimSuffix(s, " bbs")
	case "FTP":
		s = strings.TrimSuffix(s, " ftp")
	}

	// Replace punctuation with spaces, but collapse apostrophes inside words (e.g. edie's -> edies)
	s = strings.ReplaceAll(s, "'s", "s")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}

	rawWords := strings.Fields(b.String())
	filtered := make([]string, 0, len(rawWords))
	for _, w := range rawWords {
		w = normalizeNumeral(w)
		if isStopWord(w) {
			continue
		}
		filtered = append(filtered, w)
	}
	// If all words were stop words (e.g. "The Group"), fall back to rawWords without "the"
	if len(filtered) == 0 {
		for _, w := range rawWords {
			if w != "the" {
				filtered = append(filtered, normalizeNumeral(w))
			}
		}
	}

	fingerprint = strings.Join(filtered, "")
	core = fingerprint

	if len(filtered) > 1 {
		sorted := slices.Clone(filtered)
		slices.Sort(sorted)
		sortedWords = strings.Join(sorted, " ")
	}
	return fingerprint, sortedWords, core
}

// isSingleTypo returns true if a and b differ by exactly 1 edit (insertion, deletion,
// substitution, or adjacent transposition), and that edit does not involve a numeric digit.
func isSingleTypo(a, b string) bool {
	la, lb := len(a), len(b)
	if la == lb {
		// Check substitution or adjacent transposition
		diffs := make([]int, 0, 3)
		for i := range la {
			if a[i] != b[i] {
				diffs = append(diffs, i)
				if len(diffs) > 2 {
					return false
				}
			}
		}
		if len(diffs) == 1 {
			idx := diffs[0]
			if isDigit(a[idx]) || isDigit(b[idx]) {
				return false
			}
			return true
		}
		if len(diffs) == 2 && diffs[1] == diffs[0]+1 {
			i, j := diffs[0], diffs[1]
			if a[i] == b[j] && a[j] == b[i] {
				if isDigit(a[i]) || isDigit(a[j]) {
					return false
				}
				return true
			}
		}
		return false
	}

	// Ensure a is the shorter string (len(b) == len(a) + 1)
	if la > lb {
		a, b = b, a
		la, lb = lb, la
	}
	if lb-la != 1 {
		return false
	}

	i, j := 0, 0
	var skipped byte
	diffs := 0
	for i < la && j < lb {
		if a[i] == b[j] {
			i++
			j++
			continue
		}
		diffs++
		if diffs > 1 {
			return false
		}
		skipped = b[j]
		j++
	}
	if diffs == 0 {
		skipped = b[lb-1]
	}
	if isDigit(skipped) {
		return false
	}
	// Avoid flagging simple plural 's' at the very end if short, or allow it as a valid duplicate candidate
	return true
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
