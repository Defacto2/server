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

const (
	xbbs = "BBS"
	xftp = "FTP"
)

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
	visitDupEntries(tiers, func(e *DupEntry) {
		if e.Count > 0 && e.Count <= maxMinorFiles && e.RawName != "" {
			needed[e.RawName] = struct{}{}
		}
	})
	if len(needed) == 0 {
		return nil
	}

	fileRows, err := queryMinorFiles(ctx, exec, needed)
	if err != nil {
		return err
	}

	byName := groupMinorFiles(fileRows, needed)
	visitDupEntries(tiers, func(e *DupEntry) {
		if files, ok := byName[e.RawName]; ok {
			e.Files = files
		}
	})
	return nil
}

func visitDupEntries(tiers [][]DupGroup, fn func(*DupEntry)) {
	for _, groups := range tiers {
		for i := range groups {
			for j := range groups[i].Entries {
				fn(&groups[i].Entries[j])
			}
		}
	}
}

func queryMinorFiles(ctx context.Context, exec boil.ContextExecutor, needed map[string]struct{}) ([]dupFileRow, error) {
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
		return nil, err
	}
	return fileRows, nil
}

func groupMinorFiles(fileRows []dupFileRow, needed map[string]struct{}) map[string][]DupFile {
	byName := make(map[string][]DupFile, len(needed))
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
			appendMinorFile(byName, needed, fr.GroupBrandFor.String, df)
		}
		if fr.GroupBrandBy.Valid {
			appendMinorFile(byName, needed, fr.GroupBrandBy.String, df)
		}
	}
	return byName
}

func appendMinorFile(byName map[string][]DupFile, needed map[string]struct{}, name string, f DupFile) {
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

type parsedItem struct {
	entry       DupEntry
	fingerprint string
	sortedWords string
	core        string
}

type pairTracker map[string]struct{}

func (p pairTracker) has(a, b string) bool {
	_, ok := p[pairKey(a, b)]
	return ok
}

func (p pairTracker) markCluster(entries []DupEntry) {
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			p[pairKey(entries[i].Name, entries[j].Name)] = struct{}{}
		}
	}
}

func (p pairTracker) allSeen(entries []DupEntry) bool {
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			if !p.has(entries[i].Name, entries[j].Name) {
				return false
			}
		}
	}
	return true
}

// AnalyzeDuplicates runs the 3-tier in-memory clustering pipeline on a slice of releaser counts.
func AnalyzeDuplicates(rows []ReleaserCount) *DupReport {
	const (
		minRows    = 2
		minFP      = 3
		tier1      = 1
		tier2      = 2
		tier1Label = "Normalized match"
		tier2Label = "Word / token swap"
	)
	report := &DupReport{
		TotalReleasers: len(rows),
	}
	if len(rows) < minRows {
		return report
	}

	items := parseDupItems(rows)
	seen := make(pairTracker)

	report.Tier1 = clusterByKey(items, seen, tier1, tier1Label, false, func(it parsedItem) string {
		if len(it.fingerprint) < minFP {
			return ""
		}
		return it.fingerprint
	})
	report.Tier2 = clusterByKey(items, seen, tier2, tier2Label, true, func(it parsedItem) string {
		if !strings.Contains(it.sortedWords, " ") {
			return ""
		}
		return it.sortedWords
	})
	report.Tier3 = clusterTypos(items, seen)

	report.TotalClusters = len(report.Tier1) + len(report.Tier2) + len(report.Tier3)
	return report
}

func parseDupItems(rows []ReleaserCount) []parsedItem {
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
				URI:     helper.Slug(raw),
				Kind:    kind,
				Count:   r.Count,
			},
			fingerprint: fp,
			sortedWords: sortedW,
			core:        core,
		})
	}
	return items
}

func clusterByKey(
	items []parsedItem, seen pairTracker, tier int, reason string, skipSeen bool, keyFn func(parsedItem) string,
) []DupGroup {
	const minGroup = 2
	buckets := make(map[string][]DupEntry)
	order := make([]string, 0)
	for _, it := range items {
		sub := keyFn(it)
		if sub == "" {
			continue
		}
		k := it.entry.Kind + ":" + sub
		if _, exists := buckets[k]; !exists {
			order = append(order, k)
		}
		buckets[k] = append(buckets[k], it.entry)
	}

	var groups []DupGroup
	for _, k := range order {
		group := buckets[k]
		if len(group) < minGroup || (skipSeen && seen.allSeen(group)) {
			continue
		}
		sortEntries(group)
		seen.markCluster(group)
		groups = append(groups, DupGroup{
			Key:     strings.TrimPrefix(k, group[0].Kind+":"),
			Kind:    group[0].Kind,
			Tier:    tier,
			Reason:  reason,
			Entries: group,
		})
	}
	return groups
}

func clusterTypos(items []parsedItem, seen pairTracker) []DupGroup {
	const (
		minCoreLen = 6
		prefixLen  = 2
		minBucket  = 2
	)
	buckets := make(map[string][]parsedItem)
	order := make([]string, 0)
	for _, it := range items {
		if len(it.core) < minCoreLen {
			continue
		}
		pk := it.entry.Kind + ":" + it.core[:prefixLen]
		if _, exists := buckets[pk]; !exists {
			order = append(order, pk)
		}
		buckets[pk] = append(buckets[pk], it)
	}

	var groups []DupGroup
	for _, pk := range order {
		bucket := buckets[pk]
		if len(bucket) < minBucket {
			continue
		}
		groups = append(groups, bucketTypos(bucket, seen)...)
	}
	return groups
}

func bucketTypos(bucket []parsedItem, seen pairTracker) []DupGroup {
	const tier3 = 3
	var groups []DupGroup
	for i := range bucket {
		a := bucket[i]
		for j := i + 1; j < len(bucket); j++ {
			b := bucket[j]
			if a.core == b.core || !singleTypo(a.core, b.core) {
				continue
			}
			if seen.has(a.entry.Name, b.entry.Name) {
				continue
			}
			pair := []DupEntry{a.entry, b.entry}
			sortEntries(pair)
			seen.markCluster(pair)
			groups = append(groups, DupGroup{
				Key:     a.core + " ~ " + b.core,
				Kind:    a.entry.Kind,
				Tier:    tier3,
				Reason:  "1-char typo / swap",
				Entries: pair,
			})
		}
	}
	return groups
}

func classifyKind(name string) string {
	u := strings.ToUpper(strings.TrimSpace(name))
	switch {
	case strings.HasSuffix(u, " BBS") || u == xbbs:
		return xbbs
	case strings.HasSuffix(u, " FTP") || u == xftp:
		return xftp
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
	case xbbs:
		s = strings.TrimSuffix(s, " bbs")
	case xftp:
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
