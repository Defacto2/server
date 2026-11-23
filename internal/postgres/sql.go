package postgres

// Package file sql.go contains custom SQL statements that cannot be created using the SQLBoiler tool.

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/releaser"
)

type (
	SQL     string // SQL is a raw query statement for PostgreSQL.
	Version string // Version of the PostgreSQL database server in use.
	Role    string // Role is a scener attribution used in the database.
)

const (
	Writer   Role = "(upper(credit_text))"         // Writer or author of a document.
	Artist   Role = "(upper(credit_illustration))" // Artist or illustrator of an image or artwork.
	Coder    Role = "(upper(credit_program))"      // Coder or programmer of a program or application.
	Musician Role = "(upper(credit_audio))"        // Musician or composer of a music or audio track.
)

const (
	// TotalCnt is a partial SQL statement to count the number of records.
	TotalCnt = "COUNT(*) AS count_total"
	// SumSize is a partial SQL statement to sum the file size values of multiple records.
	SumSize = "SUM(filesize) AS size_total"
	// MinYear is a partial SQL statement to select the minimum year value.
	MinYear = "MIN(date_issued_year) AS min_year"
	// MaxYear is a partial SQL statement to select the maximum year value.
	MaxYear = "MAX(date_issued_year) AS max_year"
	// SelVersion is a SQL statement to select the version of the PostgreSQL database server in use.
	SelVersion = "SELECT version();"
)

// Query the database version.
func (v *Version) Query(db *sql.DB) error {
	if db == nil {
		return nil
	}

	const format = "postgres version query %s: %w"
	if err := db.QueryRow(SelVersion).Scan(v); err != nil { //nolint:noctx
		return fmt.Errorf(format, "scan", err)
	}

	return nil
}

func (v *Version) String() string {
	const (
		maxParts     = 3
		minParts     = 2
		floatBitSize = 32
	)
	s := string(*v)
	parts := strings.SplitN(s, " ", maxParts)
	if len(parts) > minParts {
		_, err := strconv.ParseFloat(parts[1], floatBitSize)
		if err != nil {
			return s
		}
		return "and using " + parts[0] + " " + parts[1]
	}
	return s
}

const (
	// releaserSEL is a partial SQL statement to select the releasers name, file count and file size sum.
	releaserSEL SQL = distReleaser +
		countSum +
		sumSize +
		fromFiles +
		combineGroup +
		whereRelNull

	// releaserForSEL is a partial SQL statement to select the releasers for name, file count, file size sum,
	// and is used for magazine titles.
	releaserForSEL SQL = distReleaser +
		countSum +
		sumSize +
		fromFiles +
		onlyGroup +
		whereRelNull

	// all SQL statements must have a tailing space,
	// otherwise the statement can break when concatenated with other SQL statements.

	fromFiles SQL = "FROM files "
	// select distinct releaser names.
	distReleaser SQL = "SELECT DISTINCT releaser, "
	// count the number of files per releaser.
	countSum SQL = "COUNT(files.filename) AS count_sum, "
	// exclude empty releaser names.
	whereRelNull SQL = "WHERE NULLIF(releaser, '') IS NOT NULL "
	// only include releasers with public records.
	deletedatIsNull SQL = "AND files.deletedat IS NULL "
	// order the list by the oldest year and the releaser name.
	orderReleaser SQL = "ORDER BY releaser ASC "
	// require the releaser name to be not empty.
	havingCount    SQL = "HAVING (COUNT(files.filename) > 0) "
	havingCountTwo SQL = "HAVING (COUNT(files.filename) > 1) "
	// group the results by the releaser name.
	groupbyRel SQL = "GROUP BY releaser "
	// exclude BBS and FTP sites.
	bbsFTPRel SQL = "AND releaser !~ 'BBS\\M' AND releaser !~ 'FTP\\M' "
	// require BBS sites.
	bbsRel SQL = "AND releaser ~ 'BBS\\M' "
	// require magazines.
	magazine SQL = "AND section = 'magazine' "
	// select the oldest year per releaser.
	minYear SQL = "MIN(files.date_issued_year) AS min_year "
	// combine the group_brand_for and group_brand_by columns as releasers.
	combineGroup SQL = "CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) "
	// only group is a replacement for combineGroup when only the group_brand_for column
	// is required such as for magazines titles.
	onlyGroup = SQL("CROSS JOIN LATERAL (values(group_brand_for)) AS T(releaser) ")
	// filter the results by the file count and the oldest year,
	// this is to exclude releasers with less than 1 file or an unknown release year.
	filterCountAndYear SQL = "HAVING (COUNT(files.filename) > 0) AND (MIN(files.date_issued_year) > 1970) "
	// order the list by the oldest year and the releaser name.
	orderSumDesc SQL = "ORDER BY count_sum DESC, releaser ASC "
	// order the list by the oldest year and the releaser name.
	orderMinYearRel SQL = "ORDER BY min_year ASC, releaser ASC "
	// sum the filesize of files per releaser.
	sumSize      SQL = "SUM(files.filesize) AS size_total "
	sumSizeComma SQL = "SUM(files.filesize) AS size_total, "
)

// ReleasersAlphabetical selects a list of distinct releasers or groups,
// excluding BBS and FTP sites.
func ReleasersAlphabetical() SQL {
	return releaserSEL +
		deletedatIsNull +
		bbsFTPRel +
		groupbyRel +
		havingCount +
		orderReleaser
}

// ReleasersAlphabeticalTwo selects a list of distinct releasers or groups with 2 or more artifacts,
// excluding BBS and FTP sites.
func ReleasersAlphabeticalTwo() SQL {
	return releaserSEL +
		deletedatIsNull +
		bbsFTPRel +
		groupbyRel +
		havingCountTwo +
		orderReleaser
}

// BBSsAlphabetical selects a list of distinct BBS names.
func BBSsAlphabetical() SQL {
	return releaserSEL +
		deletedatIsNull +
		bbsRel + // require BBS sites
		groupbyRel +
		havingCount +
		orderReleaser
}

// BBSsAlphabeticalTwo selects a list of distinct BBS names with 2 or more artifacts.
func BBSsAlphabeticalTwo() SQL {
	return releaserSEL +
		deletedatIsNull +
		bbsRel + // require BBS sites
		groupbyRel +
		havingCountTwo +
		orderReleaser
}

// FTPsAlphabetical selects a list of distinct FTP site names.
func FTPsAlphabetical() SQL {
	return releaserSEL +
		deletedatIsNull +
		"AND releaser ~ 'FTP\\M' " + // require FTP sites
		groupbyRel +
		havingCount +
		orderReleaser
}

// MagazinesAlphabetical selects a list of distinct magazine titles.
func MagazinesAlphabetical() SQL {
	return releaserForSEL + // we previously used: releaserSEL +
		deletedatIsNull +
		magazine +
		groupbyRel +
		havingCount +
		orderReleaser
}

// ReleasersProlific selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the file count (2 or more artifacts).
func ReleasersProlific() SQL {
	return releaserSEL +
		deletedatIsNull +
		bbsFTPRel +
		groupbyRel +
		havingCountTwo +
		orderSumDesc
}

// BBSsProlific selects a list of distinct releasers or groups,
// only showing BBS sites with 2 or more artifacts and ordered by the file count.
func BBSsProlific() SQL {
	return releaserSEL +
		deletedatIsNull +
		bbsRel + // require BBS sites
		groupbyRel +
		havingCountTwo +
		orderSumDesc
}

// ReleasersOldest selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the oldest year.
func ReleasersOldest() SQL {
	return distReleaser +
		countSum +
		sumSizeComma +
		minYear +
		fromFiles +
		combineGroup +
		whereRelNull +
		deletedatIsNull +
		bbsFTPRel +
		groupbyRel +
		filterCountAndYear +
		orderMinYearRel
}

// BBSsOldest selects a list of distinct releasers or groups,
// only showing BBS sites and ordered by the file count.
func BBSsOldest() SQL {
	return distReleaser +
		countSum +
		sumSizeComma +
		minYear +
		fromFiles +
		combineGroup +
		whereRelNull +
		deletedatIsNull +
		bbsRel +
		groupbyRel +
		filterCountAndYear +
		orderMinYearRel
}

// MagazinesOldest selects a list of distinct releasers or groups,
// only showing magazines and ordered by the file count.
func MagazinesOldest() SQL {
	return distReleaser +
		countSum +
		sumSizeComma +
		minYear +
		fromFiles +
		onlyGroup + // we previously used: combineGroup +
		whereRelNull +
		deletedatIsNull +
		magazine +
		groupbyRel +
		filterCountAndYear +
		orderMinYearRel
}

// ScenerSQL returns a parameterized SQL query for getting sceners and their parameters.
// This prevents SQL injection by separating the query structure from user input.
func ScenerSQL(name string) (string, []any) {
	n := strings.ToUpper(releaser.Humanize(name))
	escaped := escapeLike(n)
	// $1 is the exact uppercase handle; $2 is LIKE-escaped for comma-delimited list matching.
	exact := "(upper(credit_text) = $1)" +
		" OR (upper(credit_program) = $1)" +
		" OR (upper(credit_illustration) = $1)" +
		" OR (upper(credit_audio) = $1)"
	first := `(upper(credit_text) LIKE $2||',%' ESCAPE '\')` +
		` OR (upper(credit_program) LIKE $2||',%' ESCAPE '\')` +
		` OR (upper(credit_illustration) LIKE $2||',%' ESCAPE '\')` +
		` OR (upper(credit_audio) LIKE $2||',%' ESCAPE '\')`
	middle := `(upper(credit_text) LIKE '%,'||$2||',%' ESCAPE '\')` +
		` OR (upper(credit_program) LIKE '%,'||$2||',%' ESCAPE '\')` +
		` OR (upper(credit_illustration) LIKE '%,'||$2||',%' ESCAPE '\')` +
		` OR (upper(credit_audio) LIKE '%,'||$2||',%' ESCAPE '\')`
	last := `(upper(credit_text) LIKE '%,'||$2 ESCAPE '\')` +
		` OR (upper(credit_program) LIKE '%,'||$2 ESCAPE '\')` +
		` OR (upper(credit_illustration) LIKE '%,'||$2 ESCAPE '\')` +
		` OR (upper(credit_audio) LIKE '%,'||$2 ESCAPE '\')`
	return "(" + exact + ") OR (" + first + ") OR (" + middle + ") OR (" + last + ")", []any{n, escaped}
}

func Releasers() SQL {
	return "SELECT DISTINCT releaser " +
		fromFiles +
		combineGroup +
		whereRelNull +
		"GROUP BY releaser " +
		"ORDER BY releaser ASC"
}

// AllReleasersCounts selects all distinct releasers (groups, BBS, FTP) and their public file counts.
func AllReleasersCounts() SQL {
	return distReleaser +
		"COUNT(files.filename) AS count_sum " +
		fromFiles +
		combineGroup +
		whereRelNull +
		deletedatIsNull +
		"GROUP BY releaser " +
		havingCount +
		orderSumDesc
}

// SuggestReleasersCounts selects distinct releasers and counts across public and pending-approval records
// (excluding records explicitly deleted by a user).
func SuggestReleasersCounts() SQL {
	return distReleaser +
		"COUNT(files.filename) AS count_sum " +
		fromFiles +
		combineGroup +
		whereRelNull +
		"AND files.deletedby IS NULL " +
		"GROUP BY releaser " +
		havingCount +
		orderSumDesc
}

// SuggestMagazinesCounts selects distinct magazine titles and counts across public and pending-approval records.
func SuggestMagazinesCounts() SQL {
	return distReleaser +
		"COUNT(files.filename) AS count_sum " +
		fromFiles +
		onlyGroup +
		whereRelNull +
		magazine +
		"AND files.deletedby IS NULL " +
		"GROUP BY releaser " +
		havingCount +
		orderSumDesc
}

// Summary is an SQL statement to count the number of files, sum the filesize,
// select the oldest year and the newest year.
func Summary() SQL {
	return "SELECT COUNT(files.id) AS count_total, " + // count the number of files
		"SUM(files.filesize) AS size_total, " + // sum the filesize of files
		"MIN(files.date_issued_year) AS min_year, " + // select the oldest year
		"MAX(files.date_issued_year) AS max_year " + // select the newest year
		fromFiles +
		"WHERE "
}

// escapeLike escapes SQL LIKE wildcard characters using '\' as the escape character.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// likeOrClause builds "(releaser LIKE $1 ESCAPE '\' OR ...)" and matching prefix args.
func likeOrClause(like []string) (string, []any) {
	conds := make([]string, len(like))
	args := make([]any, len(like))
	for i, val := range like {
		conds[i] = "releaser LIKE $" + strconv.Itoa(i+1) + ` ESCAPE '\'`
		args[i] = escapeLike(strings.ToUpper(strings.TrimSpace(val))) + "%"
	}
	return "(" + strings.Join(conds, " OR ") + ") ", args
}

// SimilarToReleaser returns a parameterized SQL query for finding similar releasers.
func SimilarToReleaser(like ...string) (SQL, []any) {
	if len(like) == 0 {
		return "", nil
	}

	clause, args := likeOrClause(like)
	return releaserSEL +
		"AND " + SQL(clause) +
		groupbyRel +
		orderSumDesc, args
}

// SimilarToMagazine returns a parameterized SQL query for finding similar magazines.
func SimilarToMagazine(like ...string) (SQL, []any) {
	if len(like) == 0 {
		return "", nil
	}

	clause, args := likeOrClause(like)
	return releaserSEL +
		magazine +
		"AND " + SQL(clause) +
		groupbyRel +
		orderSumDesc, args
}

// SimilarToExact returns a parameterized SQL query for exact releaser matches.
func SimilarToExact(like ...string) (SQL, []any) {
	if len(like) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(like))
	args := make([]any, len(like))
	for i, val := range like {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = strings.ToUpper(strings.TrimSpace(val))
	}
	return releaserSEL +
		"AND releaser IN (" + SQL(strings.Join(placeholders, ", ")) + ") " +
		groupbyRel +
		orderSumDesc, args
}

// Roles returns all of the sceners reguardless of the attribution.
func Roles() Role {
	s := strings.Join([]string{string(Writer), string(Artist), string(Coder), string(Musician)}, ",")
	return Role(s)
}

func (r Role) Distinct() SQL {
	s := "SELECT DISTINCT ON(upper(scener)) scener " + // select distinct scener names
		string(fromFiles) +
		// combine the Role column name as sceners
		"CROSS JOIN LATERAL (values" + string(r) + ") AS T(scener) " +
		"WHERE NULLIF(scener, '') IS NOT NULL " + // exclude empty scener names
		"AND files.deletedat IS NULL " + // only include public records
		"GROUP BY scener " + // group the results by the scener name
		"HAVING (COUNT(files.filename) > 0) " +
		"ORDER BY upper(scener) ASC" // order by the scener name
	return SQL(s)
}

// Sceners selects a list of distinct sceners.
func Sceners() SQL {
	return Roles().Distinct()
}

// Writers selects a list of distinct writers.
func Writers() SQL {
	return Writer.Distinct()
}

// Artists selects a list of distinct artists.
func Artists() SQL {
	return Artist.Distinct()
}

// Coders selects a list of distinct coders.
func Coders() SQL {
	return Coder.Distinct()
}

// Musicians selects a list of distinct musicians.
func Musicians() SQL {
	return Musician.Distinct()
}

// SumSection is an SQL statement to sum the filesizes of public records matching the section.
func SumSection() SQL {
	return "SELECT COALESCE(SUM(files.filesize), 0) AS size_total FROM files WHERE section = $1 AND deletedat IS NULL"
}

// SumGroup is an SQL statement to sum the filesizes of public records matching the group.
func SumGroup() SQL {
	return "SELECT COALESCE(SUM(filesize), 0) AS size_total FROM files WHERE (group_brand_for = $1 OR group_brand_by = $1) AND deletedat IS NULL"
}

// SumPlatform is an SQL statement to sum the filesizes of public records matching the platform.
func SumPlatform() SQL {
	return "SELECT COALESCE(SUM(filesize), 0) AS size_total FROM files WHERE platform = $1 AND deletedat IS NULL"
}

// SetUpper is an SQL statement to update a column with trimmed uppercase values.
func SetUpper(column string) string {
	return fmt.Sprintf(
		"UPDATE files SET %s = UPPER(BTRIM(%s)) WHERE %s IS NOT NULL AND %s::text <> UPPER(BTRIM(%s::text));",
		column, column, column, column, column,
	)
}

// SetFilesize0 is an SQL statement to update filesize column NULLs with 0 values.
// This is a fix for the error: failed to bind pointers to obj: sql:
// Scan error on column index 2, name "size_total": converting NULL to int is unsupported.
func SetFilesize0() string {
	return "UPDATE files SET filesize = 0 WHERE filesize IS NULL;"
}

// NumSuffixes is an SQL statement to list files with numeric suffixes.
func NumSuffixes() string {
	return `SELECT id, uuid, filename AS suffix
		FROM files
		WHERE filename ~ '\([0-9]{1,3}\)\.[A-Za-z0-9]+$'
		AND filename !~ '\([12][0-9]{3}\)'
		ORDER BY filename
		LIMIT 100;`
}
