package postgres

import (
	"slices"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

// TestVersionString tests the Version.String() method with various inputs.
func TestVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name:     "empty version",
			version:  Version(""),
			expected: "",
		},
		{
			name:     "single word",
			version:  Version("PostgreSQL"),
			expected: "PostgreSQL",
		},
		{
			name:     "valid version with 3+ parts",
			version:  Version("PostgreSQL 13.8 on x86_64-pc-linux-gnu"),
			expected: "and using PostgreSQL 13.8",
		},
		{
			name:     "version with non-numeric second part",
			version:  Version("PostgreSQL alpha on x86_64"),
			expected: "PostgreSQL alpha on x86_64",
		},
		{
			name:     "short version string with 2 parts",
			version:  Version("PostgreSQL 14"),
			expected: "PostgreSQL 14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			be.Equal(t, tt.expected, result)
		})
	}
}

// TestColumns verifies the Columns function returns expected column selections.
func TestColumns(t *testing.T) {
	cols := Columns()
	be.Equal(t, 4, len(cols))
	be.Equal(t, SumSize, cols[0])
	be.Equal(t, TotalCnt, cols[1])
	be.Equal(t, MinYear, cols[2])
	be.Equal(t, MaxYear, cols[3])
}

// TestStat verifies the Stat function returns expected column selections.
func TestStat(t *testing.T) {
	stats := Stat()
	be.Equal(t, 2, len(stats))
	be.Equal(t, SumSize, stats[0])
	be.Equal(t, TotalCnt, stats[1])
}

// TestReleasersAlphabetical verifies the SQL query construction.
func TestReleasersAlphabetical(t *testing.T) {
	query := ReleasersAlphabetical()
	queryStr := string(query)

	// Should contain key SQL components
	be.True(t, strings.Contains(queryStr, "SELECT DISTINCT releaser"))
	be.True(t, strings.Contains(queryStr, "FROM files"))
	be.True(t, strings.Contains(queryStr, "WHERE NULLIF(releaser, '') IS NOT NULL"))
	be.True(t, strings.Contains(queryStr, "BBS")) // Exclude BBS and FTP
	be.True(t, strings.Contains(queryStr, "ORDER BY releaser ASC"))
}

// TestBBSsAlphabetical verifies BBS sites query construction.
func TestBBSsAlphabetical(t *testing.T) {
	query := BBSsAlphabetical()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "BBS"))
	be.True(t, strings.Contains(queryStr, "ORDER BY releaser ASC"))
}

// TestMagazinesAlphabetical verifies magazines query construction.
func TestMagazinesAlphabetical(t *testing.T) {
	query := MagazinesAlphabetical()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "magazine"))
	be.True(t, strings.Contains(queryStr, "ORDER BY releaser ASC"))
}

// TestReleasersProlific verifies prolific releasers query ordering.
func TestReleasersProlific(t *testing.T) {
	query := ReleasersProlific()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "ORDER BY count_sum DESC"))
}

// TestReleasersOldest verifies oldest releasers query construction.
func TestReleasersOldest(t *testing.T) {
	query := ReleasersOldest()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "MIN(files.date_issued_year)"))
	be.True(t, strings.Contains(queryStr, "ORDER BY min_year ASC"))
}

// TestScenerSQL tests parameterized query construction for sceners.
func TestScenerSQL(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		expectedCount int
		shouldBeEmpty bool
	}{
		{
			name:          "simple name",
			input:         []string{"john"},
			expectedCount: 1,
		},
		{
			name:          "name with spaces",
			input:         []string{"  john doe  "},
			expectedCount: 1,
		},
		{
			name:          "empty string",
			input:         []string{""},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := ScenerSQL(tt.input[0])

			// Query should not be empty
			be.True(t, len(query) > 0)

			// Should have exactly one parameter
			be.Equal(t, tt.expectedCount, len(params))

			// Query should contain parameterized placeholder
			be.True(t, strings.Contains(query, "$1"))

			// Query should contain OR conditions for credit types
			be.True(t, strings.Contains(query, "credit_text"))
			be.True(t, strings.Contains(query, "credit_program"))
			be.True(t, strings.Contains(query, "credit_illustration"))
			be.True(t, strings.Contains(query, "credit_audio"))
		})
	}
}

// TestSimilarToReleaser tests parameterized query for similar releasers.
func TestSimilarToReleaser(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		expectedCount int
		shouldBeEmpty bool
	}{
		{
			name:          "single releaser",
			input:         []string{"Lotus"},
			expectedCount: 1,
			shouldBeEmpty: false,
		},
		{
			name:          "multiple releasers",
			input:         []string{"Lotus", "Amiga"},
			expectedCount: 2,
			shouldBeEmpty: false,
		},
		{
			name:          "empty input",
			input:         []string{},
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "single releaser with whitespace",
			input:         []string{"  Lotus  "},
			expectedCount: 1,
			shouldBeEmpty: false,
		},
		{
			name:          "duplicate releasers",
			input:         []string{"Lotus", "lotus", "LOTUS"},
			expectedCount: 3, // All converted to uppercase
			shouldBeEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := SimilarToReleaser(tt.input...)

			if tt.shouldBeEmpty {
				be.Equal(t, "", string(query))
				be.Equal(t, 0, len(params))
			} else {
				// Query should contain SIMILAR TO
				be.True(t, strings.Contains(string(query), "SIMILAR TO"))

				// Should always have exactly 1 parameter (the combined pattern)
				be.Equal(t, 1, len(params))

				// Parameter should be a string
				val, ok := params[0].(string)
				be.True(t, ok)

				// Parameter should contain the joined values with | separator
				be.True(t, strings.Contains(val, "|") || len(tt.input) == 1)

				// Query should have $1 placeholder
				be.True(t, strings.Contains(string(query), "$1"))
			}
		})
	}
}

// TestSimilarToMagazine tests parameterized query for similar magazines.
func TestSimilarToMagazine(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		shouldBeEmpty bool
	}{
		{
			name:          "single magazine",
			input:         []string{"Amiga World"},
			shouldBeEmpty: false,
		},
		{
			name:          "multiple magazines",
			input:         []string{"Amiga World", "PC Zone"},
			shouldBeEmpty: false,
		},
		{
			name:          "empty input",
			input:         []string{},
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := SimilarToMagazine(tt.input...)

			if tt.shouldBeEmpty {
				be.Equal(t, "", string(query))
				be.Equal(t, 0, len(params))
			} else {
				// Query should contain magazine filter
				be.True(t, strings.Contains(string(query), "magazine"))
				be.True(t, strings.Contains(string(query), "SIMILAR TO"))
				// Should always have 1 parameter (combined pattern)
				be.Equal(t, 1, len(params))
			}
		})
	}
}

// TestSimilarToExact tests parameterized query for exact matches.
func TestSimilarToExact(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		shouldBeEmpty bool
	}{
		{
			name:          "single exact match",
			input:         []string{"Breadbox"},
			shouldBeEmpty: false,
		},
		{
			name:          "multiple exact matches",
			input:         []string{"Breadbox", "Apogee"},
			shouldBeEmpty: false,
		},
		{
			name:          "empty input",
			input:         []string{},
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := SimilarToExact(tt.input...)

			if tt.shouldBeEmpty {
				be.Equal(t, "", string(query))
				be.Equal(t, 0, len(params))
			} else {
				be.True(t, strings.Contains(string(query), "SIMILAR TO"))
				// Should always have 1 parameter (combined pattern)
				be.Equal(t, 1, len(params))
			}
		})
	}
}

// TestSimilarToReleaser_ParameterValidation verifies parameters are properly formatted.
func TestSimilarToReleaser_ParameterValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "trim whitespace",
			input:    []string{"  Space  "},
			expected: "SPACE",
		},
		{
			name:     "convert to uppercase",
			input:    []string{"lowercase"},
			expected: "LOWERCASE",
		},
		{
			name:     "multiple values joined with pipe",
			input:    []string{"first", "  second  ", "THIRD"},
			expected: "FIRST|SECOND|THIRD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, params := SimilarToReleaser(tt.input...)

			// Should always have 1 parameter
			be.Equal(t, 1, len(params))

			val := params[0].(string)
			be.Equal(t, tt.expected, val)
		})
	}
}

// TestRolesConstants verifies Role type constants are defined correctly.
func TestRolesConstants(t *testing.T) {
	be.Equal(t, Role("(upper(credit_text))"), Writer)
	be.Equal(t, Role("(upper(credit_program))"), Coder)
	be.Equal(t, Role("(upper(credit_illustration))"), Artist)
	be.Equal(t, Role("(upper(credit_audio))"), Musician)
}

// TestRoles verifies the Roles function returns all roles joined.
func TestRoles(t *testing.T) {
	roles := Roles()
	roleStr := string(roles)

	be.True(t, strings.Contains(roleStr, string(Writer)))
	be.True(t, strings.Contains(roleStr, string(Artist)))
	be.True(t, strings.Contains(roleStr, string(Coder)))
	be.True(t, strings.Contains(roleStr, string(Musician)))
	be.True(t, strings.Contains(roleStr, ","))
}

// TestRoleDistinct verifies the Distinct method constructs valid SQL.
func TestRoleDistinct(t *testing.T) {
	role := Writer
	query := role.Distinct()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "SELECT DISTINCT ON"))
	be.True(t, strings.Contains(queryStr, "scener"))
	be.True(t, strings.Contains(queryStr, "FROM files"))
	be.True(t, strings.Contains(queryStr, "CROSS JOIN LATERAL"))
	be.True(t, strings.Contains(queryStr, "ORDER BY upper(scener) ASC"))
}

// TestScenersFunctions verifies scener query functions.
func TestScenersFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() SQL
		roleFind string
	}{
		{
			name:     "Sceners",
			fn:       Sceners,
			roleFind: "credit_text,",
		},
		{
			name:     "Writers",
			fn:       Writers,
			roleFind: "credit_text",
		},
		{
			name:     "Artists",
			fn:       Artists,
			roleFind: "credit_illustration",
		},
		{
			name:     "Coders",
			fn:       Coders,
			roleFind: "credit_program",
		},
		{
			name:     "Musicians",
			fn:       Musicians,
			roleFind: "credit_audio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.fn()
			queryStr := string(query)
			be.True(t, strings.Contains(queryStr, "SELECT DISTINCT"))
			be.True(t, strings.Contains(queryStr, "scener"))
		})
	}
}

// TestSumSection verifies SumSection query.
func TestSumSection(t *testing.T) {
	query := SumSection()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "SUM(files.filesize)"))
	be.True(t, strings.Contains(queryStr, "section = $1"))
}

// TestSumGroup verifies SumGroup query.
func TestSumGroup(t *testing.T) {
	query := SumGroup()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "SUM(filesize)"))
	be.True(t, strings.Contains(queryStr, "group_brand_for = $1"))
}

// TestSumPlatform verifies SumPlatform query.
func TestSumPlatform(t *testing.T) {
	query := SumPlatform()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "sum(filesize)"))
	be.True(t, strings.Contains(queryStr, "platform = $1"))
}

// TestSetUpper verifies SetUpper query construction.
func TestSetUpper(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected string
	}{
		{
			name:     "releaser column",
			column:   "releaser",
			expected: "UPDATE files SET releaser = UPPER(releaser);",
		},
		{
			name:     "record_title column",
			column:   "record_title",
			expected: "UPDATE files SET record_title = UPPER(record_title);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := SetUpper(tt.column)
			be.Equal(t, tt.expected, query)
		})
	}
}

// TestSetFilesize0 verifies SetFilesize0 query.
func TestSetFilesize0(t *testing.T) {
	query := SetFilesize0()
	expected := "UPDATE files SET filesize = 0 WHERE filesize IS NULL;"
	be.Equal(t, expected, query)
}

// TestSummary verifies Summary query construction.
func TestSummary(t *testing.T) {
	query := Summary()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "COUNT(files.id)"))
	be.True(t, strings.Contains(queryStr, "SUM(files.filesize)"))
	be.True(t, strings.Contains(queryStr, "MIN(files.date_issued_year)"))
	be.True(t, strings.Contains(queryStr, "MAX(files.date_issued_year)"))
	be.True(t, strings.Contains(queryStr, "FROM files"))
}

// TestReleasers verifies Releasers query construction.
func TestReleasers(t *testing.T) {
	query := Releasers()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "SELECT DISTINCT releaser"))
	be.True(t, strings.Contains(queryStr, "GROUP BY releaser"))
	be.True(t, strings.Contains(queryStr, "ORDER BY releaser ASC"))
}

// TestBBSsOldest verifies BBS oldest query construction.
func TestBBSsOldest(t *testing.T) {
	query := BBSsOldest()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "BBS"))
	be.True(t, strings.Contains(queryStr, "MIN(files.date_issued_year)"))
	be.True(t, strings.Contains(queryStr, "ORDER BY min_year ASC"))
}

// TestMagazinesOldest verifies magazines oldest query construction.
func TestMagazinesOldest(t *testing.T) {
	query := MagazinesOldest()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "magazine"))
	be.True(t, strings.Contains(queryStr, "MIN(files.date_issued_year)"))
	be.True(t, strings.Contains(queryStr, "ORDER BY min_year ASC"))
}

// TestFTPsAlphabetical verifies FTP sites query construction.
func TestFTPsAlphabetical(t *testing.T) {
	query := FTPsAlphabetical()
	queryStr := string(query)

	be.True(t, strings.Contains(queryStr, "FTP"))
	be.True(t, strings.Contains(queryStr, "ORDER BY releaser ASC"))
}

// TestSimilarToFunctions_NoInputCloning verifies input slices are not mutated.
func TestSimilarToFunctions_NoInputCloning(t *testing.T) {
	tests := []struct {
		name string
		fn   func(...string) (SQL, []any)
	}{
		{
			name: "SimilarToReleaser",
			fn:   SimilarToReleaser,
		},
		{
			name: "SimilarToMagazine",
			fn:   SimilarToMagazine,
		},
		{
			name: "SimilarToExact",
			fn:   SimilarToExact,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []string{"test", "  space  "}
			inputCopy := slices.Clone(input)

			_, _ = tt.fn(input...)

			// Original input should not be modified
			be.Equal(t, inputCopy, input)
		})
	}
}

// TestSimilarToPlaceholderFormat verifies placeholder format is correct.
func TestSimilarToPlaceholderFormat(t *testing.T) {
	tests := []struct {
		name              string
		input             []string
		expectedPattern   string
		placeholderFormat string
	}{
		{
			name:            "single placeholder",
			input:           []string{"one"},
			expectedPattern: "$1",
		},
		{
			name:            "two placeholders",
			input:           []string{"one", "two"},
			expectedPattern: "$1",
		},
		{
			name:            "three placeholders",
			input:           []string{"one", "two", "three"},
			expectedPattern: "$1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, _ := SimilarToReleaser(tt.input...)
			queryStr := string(query)

			// All queries should have $1 since we now use a single parameter
			be.True(t, strings.Contains(queryStr, tt.expectedPattern))
		})
	}
}

// BenchmarkScenerSQL benchmarks the ScenerSQL function.
func BenchmarkScenerSQL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ScenerSQL("John Doe")
	}
}

// BenchmarkSimilarToReleaser benchmarks the SimilarToReleaser function.
func BenchmarkSimilarToReleaser(b *testing.B) {
	input := []string{"Lotus", "Amiga", "Test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SimilarToReleaser(input...)
	}
}

// BenchmarkSimilarToMagazine benchmarks the SimilarToMagazine function.
func BenchmarkSimilarToMagazine(b *testing.B) {
	input := []string{"Amiga World", "PC Zone"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SimilarToMagazine(input...)
	}
}

// BenchmarkSimilarToExact benchmarks the SimilarToExact function.
func BenchmarkSimilarToExact(b *testing.B) {
	input := []string{"Breadbox", "Apogee", "Lotus"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SimilarToExact(input...)
	}
}
