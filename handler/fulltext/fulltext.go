// Package fulltext uses a full-text search engine to index,
// and search the content of the markdown textfiles used within
// the application.
//
// Currently, the index is built at startup and stored in RAM as the
// number of files to catalog is low. However, if we were to expand
// the indexing to file downloads a different approach is needed.
package fulltext

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strconv"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
	"github.com/wizenheimer/blaze"
	stripmd "github.com/writeas/go-strip-markdown"
)

var (
	ErrNoBody  = errors.New("body is empty or only contains white space")
	ErrNoIndex = errors.New("the blaze index is empty and must be created before using this func")
	ErrNoName  = errors.New("filename is empty")
)

const Window = 40 // Window is number of characters to display either side of a snippet

type Index struct {
	Name string // Name of the indexed file
	Body string // Body or text content of the indexed file
}

type Result struct {
	ID    int     // ID of the matching file
	Name  string  // Name of the matching file
	Score float64 // Score is the match relevancy
	Snip  string  // Snip is an snippet of text that was matched
}

// Tidbits are the index for the group biographies stored as markdown texts.
type Tidbits struct {
	TotalDocs  int   // Total number of indexed documents
	TotalTerms int64 // Total number of terms across all docs
	engine     *blaze.InvertedIndex
	store      []Index
}

// Add both the filename and body to the Tidbits index.
// There are no checks for the validity of the arguments,
//
// However, errors will return when:
//   - the index has not been created
//   - the filename is empty
//   - the body trimmed of white space is empty
func (ts *Tidbits) Add(filename, body string) error {
	const name = "tidbits search add"
	if ts.engine == nil {
		return fmt.Errorf("%s: %w", name, ErrNoIndex)
	}
	if filename == "" {
		return fmt.Errorf("%s: %w", name, ErrNoName)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("%s: %w", name, ErrNoBody)
	}

	// remove any embedded html including <a href> links etc.
	htm := bluemonday.StrictPolicy()
	s := htm.Sanitize(body)
	// remove markdown styling
	s = stripmd.Strip(s)
	// remove any non-standard characters like box and line drawing chars
	s = strings.Map(filter, s)

	docID := len(ts.store)
	ts.engine.Index(docID, s)
	ts.store = append(ts.store, Index{
		Name: filename,
		Body: s,
	})

	return nil
}

// filter unwanted characters.
func filter(r rune) rune {
	switch {
	case unicode.IsLetter(r) || unicode.IsDigit(r):
		return r
	case unicode.IsSpace(r) || unicode.IsPunct(r):
		return r
	}
	return -1
}

// NewIndex indexes all the files found in the root directory of the fsys embed file system.
//
// Note, it will overwrite any existing indexing.
// The blaze library annoyingly slogs every index added,
// so this func temporary mutes all slogs while indexing.
func (ts *Tidbits) NewIndex(fsys embed.FS, root string) error {
	const name = "tidbits new index"
	ts.engine = blaze.NewInvertedIndex()

	mute := slog.New(slog.DiscardHandler)
	restore := slog.Default()
	slog.SetDefault(mute)

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return ts.Add(d.Name(), string(b))
	})
	slog.SetDefault(restore)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	ts.TotalDocs = ts.engine.TotalDocs
	ts.TotalTerms = ts.engine.TotalTerms

	return nil
}

// Search the tidbits index for the query and return the results.
// An empty Result means no results were found.
func (ts *Tidbits) Search(query string, maxResults int) []Result {
	noresult := []Result{}
	query = strings.TrimSpace(query)
	if query == "" {
		return noresult
	}

	queries := splitQuery(query)
	qb := blaze.NewQueryBuilder(ts.engine)
	if len(queries) == 1 {
		qb.Term(query)
	} else {
		for i := range queries {
			qb.And().Term(queries[i])
		}
	}
	matches := qb.ExecuteWithBM25(maxResults)
	results := make([]Result, len(matches))

	for i, match := range matches {
		name := ts.Name(match.DocID)
		body := ts.Body(match.DocID)
		results[i].ID = id(name)
		results[i].Name = name
		results[i].Score = match.Score
		results[i].Snip = Snippet(query, body, Window)
	}

	return results
}

func splitQuery(query string) []string {
	return strings.FieldsFunc(query, func(r rune) bool {
		// Split on white space, commas, or semicolons
		return r == ' ' || r == '\n' || r == ',' || r == ';'
	})
}

// id returns the named file without its markdown extension
// and as a usable int. It is intended for use as an ID value.
func id(name string) int {
	const invalid, md = -1, ".md"
	if !strings.HasSuffix(name, md) {
		return invalid
	}
	s := strings.TrimSuffix(name, md)
	i, err := strconv.Atoi(s)
	if err != nil {
		return invalid
	}
	return i
}

// Body value of the document id is returned from the index.
// If the id doesn't exist, an empty string is returned.
func (ts *Tidbits) Body(docID int) string {
	if !ts.find(docID) {
		return ""
	}
	return ts.store[docID].Body
}

// Name value of the document id is returned from the index.
// If the id doesn't exist, an "error" string is returned.
func (ts *Tidbits) Name(docID int) string {
	if !ts.find(docID) {
		return "error"
	}
	return ts.store[docID].Name
}

// New creates an empty inverted index.
func (ts *Tidbits) New() {
	ts.engine = blaze.NewInvertedIndex()
}

// Stores returns the number of items in the index store.
func (ts *Tidbits) Stores() int {
	return len(ts.store)
}

func (ts *Tidbits) find(docID int) bool {
	return docID >= 0 && docID < len(ts.store)
}

// Snippet finds the query in the body of text and returns
// the match with its surrounding text.
//
// The wordWindow is the number of words to display either
// side of the match.
//
// If no match is found, an empty string is returned.
func Snippet(query, body string, wordWindow int) string { //nolint:cyclop
	if query == "" {
		return ""
	}
	if wordWindow < 0 {
		wordWindow = 0
	}
	const eclipse = "..."

	maximum := len(body)
	lowerQuery := strings.ToLower(query)
	queryLen := len(query)
	matchStart := -1

	// brute force search is byte-safe when strings.ToLower changes string lengths
	for i := 0; i <= maximum-queryLen; i++ {
		if strings.ToLower(body[i:i+queryLen]) == lowerQuery {
			matchStart = i
			break
		}
	}

	if matchStart == -1 {
		const double = 2
		return truncateByWords(body, wordWindow*double) + eclipse
	}

	// this code is verbose because of the edge cases that came from the fuzzing tests

	matchEnd := matchStart + queryLen
	start := matchStart
	wordsBefore := 0
	for start > 0 && wordsBefore < wordWindow {
		start--
		if body[start] == ' ' || body[start] == '\n' {
			wordsBefore++
		}
	}

	end := matchEnd
	wordsAfter := 0
	for end < maximum && wordsAfter < wordWindow {
		if body[end] == ' ' || body[end] == '\n' {
			wordsAfter++
		}
		end++
	}

	if start < 0 {
		start = 0
	}
	if end > maximum {
		end = maximum
	}
	if start > end {
		start = end
	}

	snippet := strings.TrimSpace(body[start:end])

	if snippet == "" && query != "" {
		snippet = body[matchStart:matchEnd]
	}

	if start > 0 {
		snippet = eclipse + snippet
	} else if end < maximum {
		snippet += eclipse
	}

	return snippet
}

func truncateByWords(s string, maxWords int) string {
	words := strings.Fields(s)
	if len(words) > maxWords {
		return strings.Join(words[:maxWords], " ")
	}
	return s
}
