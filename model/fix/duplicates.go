//nolint:exhaustruct_v5,nonamedreturns
package fix

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
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

// ErrInvalidMerge is returned when merge source or destination releaser names are invalid.
var ErrInvalidMerge = errors.New("source and target releaser names must be non-empty and distinct")

// ReleaserCount holds a raw releaser name and its file count from the database.
type ReleaserCount struct {
	Releaser string `boil:"releaser"`
	Count    int    `boil:"count_sum"`
}

// DupFile represents an individual artifact file belonging to a minor duplicate entry.
type DupFile struct {
	ObfuscatedID string
	Filename     string
}

// DupEntry represents a single releaser candidate within a duplicate cluster.
type DupEntry struct {
	Name    string
	RawName string
	Display string
	URI     string
	Kind    string
	Count   int
	Files   []DupFile
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
	_, _ = queries.Raw(postgres.SetUpper("group_brand_for")).ExecContext(ctx, exec)
	_, _ = queries.Raw(postgres.SetUpper(groupBrandBy)).ExecContext(ctx, exec)
	var rows []ReleaserCount
	if err := queries.Raw(string(postgres.AllReleasersCounts())).Bind(ctx, exec, &rows); err != nil {
		return nil, fmt.Errorf("fix duplicates query: %w", err)
	}
	report := AnalyzeDuplicates(rows)
	if err := attachMinorFiles(ctx, exec, report); err != nil {
		return nil, fmt.Errorf("fix duplicates files: %w", err)
	}
	return report, nil
}

// MergeReleaser renames all references of releaser "from" to "to" across group_brand_for
// and group_brand_by, clearing group_brand_by if both columns become identical.
func MergeReleaser(ctx context.Context, exec boil.ContextExecutor, from, to string) (int64, error) {
	const format = "merge releaser: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, err)
	}
	if from == "" || to == "" || from == to {
		return 0, fmt.Errorf(format, ErrInvalidMerge)
	}

	resFor, err := queries.Raw(
		"UPDATE files SET group_brand_for = $1 WHERE group_brand_for = $2", to, from,
	).ExecContext(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}
	nFor, _ := resFor.RowsAffected()

	resBy, err := queries.Raw(
		"UPDATE files SET group_brand_by = $1 WHERE group_brand_by = $2", to, from,
	).ExecContext(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}
	nBy, _ := resBy.RowsAffected()

	if _, err := queries.Raw(
		"UPDATE files SET group_brand_by = '' WHERE group_brand_for = group_brand_by AND group_brand_by <> ''",
	).ExecContext(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, err)
	}

	Clear()
	return nFor + nBy, nil
}

const maxMinorFiles = 10

type dupFileRow struct {
	ID            int64          `boil:"id"`
	Filename      sql.NullString `boil:"filename"`
	GroupBrandFor sql.NullString `boil:"group_brand_for"`
	GroupBrandBy  sql.NullString `boil:"group_brand_by"`
}

// attachMinorFiles populates direct /f/:id links for any duplicate entry with <= maxMinorFiles files.
func attachMinorFiles(ctx context.Context, exec boil.ContextExecutor, report *DupReport) error {
	if report == nil {
		return nil
	}
	tiers := [][]DupGroup{report.Tier1, report.Tier2, report.Tier3}
	needed := make(map[string]struct{})
	for _, groups := range tiers {
		for _, g := range groups {
			for _, e := range g.Entries {
				if e.Count > 0 && e.Count <= maxMinorFiles && e.RawName != "" {
					needed[e.RawName] = struct{}{}
				}
			}
		}
	}
	if len(needed) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(needed))
	args := make([]any, 0, len(needed))
	idx := 1
	for name := range needed {
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		args = append(args, name)
		idx++
	}
	inClause := strings.Join(placeholders, ",")
	q := "SELECT id, filename, group_brand_for, group_brand_by FROM files " +
		"WHERE deletedat IS NULL AND (group_brand_for IN (" + inClause + ") OR group_brand_by IN (" + inClause + ")) " +
		"ORDER BY id ASC"

	var fileRows []dupFileRow
	if err := queries.Raw(q, args...).Bind(ctx, exec, &fileRows); err != nil {
		return err
	}

	byName := make(map[string][]DupFile, len(needed))
	addFile := func(name string, f DupFile) {
		if _, ok := needed[name]; !ok {
			return
		}
		if len(byName[name]) >= maxMinorFiles {
			return
		}
		for _, existing := range byName[name] {
			if existing.ObfuscatedID == f.ObfuscatedID {
				return
			}
		}
		byName[name] = append(byName[name], f)
	}
	for _, fr := range fileRows {
		fname := strings.TrimSpace(fr.Filename.String)
		if fname == "" {
			fname = fmt.Sprintf("Record #%d", fr.ID)
		}
		df := DupFile{
			ObfuscatedID: helper.ObfuscateID(fr.ID),
			Filename:     fname,
		}
		if fr.GroupBrandFor.Valid {
			addFile(fr.GroupBrandFor.String, df)
		}
		if fr.GroupBrandBy.Valid {
			addFile(fr.GroupBrandBy.String, df)
		}
	}

	for _, groups := range tiers {
		for i := range groups {
			for j := range groups[i].Entries {
				e := &groups[i].Entries[j]
				if files, ok := byName[e.RawName]; ok {
					e.Files = files
				}
			}
		}
	}
	return nil
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
				RawName: r.Releaser,
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
				if singleTypo(a.core, b.core) {
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

func stopWord(w string) bool {
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
		if stopWord(w) {
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

// singleTypo returns true if a and b differ by exactly 1 edit (insertion, deletion,
// substitution, or adjacent transposition), and that edit does not involve a numeric digit.
func singleTypo(a, b string) bool {
	switch len(b) - len(a) {
	case 0:
		return sameLenTypo(a, b)
	case 1:
		return insertTypo(a, b)
	case -1:
		return insertTypo(b, a)
	default:
		return false
	}
}

func firstDiff(a, b string) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func sameLenTypo(a, b string) bool {
	i := firstDiff(a, b)
	if i == len(a) {
		return false
	}
	if a[i+1:] == b[i+1:] {
		return !isDigit(a[i]) && !isDigit(b[i])
	}
	if i+1 < len(a) && a[i] == b[i+1] && a[i+1] == b[i] && a[i+2:] == b[i+2:] {
		return !isDigit(a[i]) && !isDigit(a[i+1])
	}
	return false
}

func insertTypo(short, long string) bool {
	i := firstDiff(short, long)
	return !isDigit(long[i]) && short[i:] == long[i+1:]
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
