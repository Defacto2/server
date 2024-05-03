package postgres

// Package file sql.go contains custom SQL statements that cannot be created using the SQLBoiler tool.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
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
	// SumSize is a partial SQL statement to sum the filesize values of multiple records.
	SumSize = "SUM(filesize) AS size_total"
	// MinYear is a partial SQL statement to select the minimum year value.
	MinYear = "MIN(date_issued_year) AS min_year"
	// MaxYear is a partial SQL statement to select the maximum year value.
	MaxYear = "MAX(date_issued_year) AS max_year"
	// Ver is a SQL statement to select the version of the PostgreSQL database server in use.
	Ver = "SELECT version();"
)

// Query the database version.
func (v *Version) Query() error {
	conn, err := ConnectDB()
	if err != nil {
		return fmt.Errorf("connectDB: %w", err)
	}
	rows, err := conn.Query(Ver)
	if err != nil {
		return fmt.Errorf("conn.Query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("rows.Err: %w", rows.Err())
	}
	defer rows.Close()
	defer conn.Close()
	for rows.Next() {
		if err := rows.Scan(v); err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}
	}
	return nil
}

func (v *Version) String() string {
	s := string(*v)
	const invalid = 2
	if x := strings.Split(s, " "); len(x) > invalid {
		_, err := strconv.ParseFloat(x[1], 32)
		if err != nil {
			return s
		}
		return "and using " + strings.Join(x[0:2], " ")
	}
	return s
}

// Columns returns a list of column selections used for filtering and statistics.
func Columns() []string {
	return []string{SumSize, TotalCnt, MinYear, MaxYear}
}

// Stat returns the SumSize and TotalCnt column selections.
func Stat() []string {
	return []string{SumSize, TotalCnt}
}

const (
	// releaserSEL is a partial SQL statement to select the releasers name, file count and filesize sum.
	releaserSEL SQL = distReleaser +
		countSum +
		sumSize +
		fromFiles +
		combineGroup +
		whereRelNull

	fromFiles SQL = "FROM files "
	// select distinct releaser names.
	distReleaser SQL = "SELECT DISTINCT releaser, "
	// count the number of files per releaser.
	countSum SQL = "COUNT(files.filename) AS count_sum, "
	// exclude empty releaser names.
	whereRelNull SQL = "WHERE NULLIF(releaser, '') IS NOT NULL "
	// releaserBy is a partial SQL statement to group the results by the releaser name.
	releaserBy SQL = "GROUP BY releaser ORDER BY releaser ASC"
	// only include releasers with public records.
	deletedatIsNull SQL = "AND files.deletedat IS NULL"
	// order the list by the oldest year and the releaser name.
	orderReleaser SQL = "ORDER BY releaser ASC"
	// require the releaser name to be not empty.
	havingCount SQL = "HAVING (COUNT(files.filename) > 0) "
	// group the results by the releaser name and deleteat.
	groupbyRel SQL = "GROUP BY releaser, files.deletedat "
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
	// filter the results by the file count and the oldest year,
	// this is to exclude releasers with less than 1 file or an unknown release year.
	filterCountAndYear SQL = "HAVING (COUNT(files.filename) > 0) AND (MIN(files.date_issued_year) > 1970) "
	// order the list by the oldest year and the releaser name.
	orderMinYearRel SQL = "ORDER BY min_year ASC, releaser ASC"
	// sum the filesize of files per releaser.
	sumSize SQL = "SUM(files.filesize) AS size_total "
)

// ReleasersAlphabetical selects a list of distinct releasers or groups,
// excluding BBS and FTP sites.
func ReleasersAlphabetical() SQL {
	return releaserSEL +
		bbsFTPRel +
		groupbyRel +
		havingCount +
		deletedatIsNull +
		orderReleaser
}

// BBSsAlphabetical selects a list of distinct BBS names.
func BBSsAlphabetical() SQL {
	return releaserSEL +
		bbsRel + // require BBS sites
		groupbyRel +
		havingCount +
		deletedatIsNull +
		orderReleaser
}

// FTPsAlphabetical selects a list of distinct FTP site names.
func FTPsAlphabetical() SQL {
	return releaserSEL +
		"AND releaser ~ 'FTP\\M' " + // require FTP sites
		groupbyRel +
		havingCount +
		deletedatIsNull +
		orderReleaser
}

// MagazinesAlphabetical selects a list of distinct magazine titles.
func MagazinesAlphabetical() SQL {
	return releaserSEL +
		magazine +
		groupbyRel +
		havingCount +
		deletedatIsNull +
		orderReleaser
}

// ReleasersProlific selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the file count.
func ReleasersProlific() SQL {
	return releaserSEL +
		bbsFTPRel +
		groupbyRel +
		havingCount +
		deletedatIsNull +
		"ORDER BY count_sum DESC, releaser ASC" // order the list by the oldest year and the releaser name
}

// BBSsProlific selects a list of distinct releasers or groups,
// only showing BBS sites and ordered by the file count.
func BBSsProlific() SQL {
	return releaserSEL +
		bbsRel + // require BBS sites
		groupbyRel +
		havingCount +
		deletedatIsNull +
		"ORDER BY count_sum DESC, releaser ASC" // order the list by the oldest year and the releaser name
}

// ReleasersOldest selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the oldest year.
func ReleasersOldest() SQL {
	return distReleaser +
		countSum +
		sumSize +
		minYear +
		fromFiles +
		combineGroup +
		whereRelNull +
		bbsFTPRel +
		groupbyRel +
		filterCountAndYear +
		deletedatIsNull +
		orderMinYearRel
}

// BBSsOldest selects a list of distinct releasers or groups,
// only showing BBS sites and ordered by the file count.
func BBSsOldest() SQL {
	return distReleaser +
		countSum +
		sumSize +
		minYear +
		fromFiles +
		combineGroup +
		whereRelNull +
		bbsRel +
		groupbyRel +
		filterCountAndYear +
		deletedatIsNull +
		orderMinYearRel
}

// MagazinesOldest selects a list of distinct releasers or groups,
// only showing magazines and ordered by the file count.
func MagazinesOldest() SQL {
	return distReleaser +
		countSum +
		sumSize +
		minYear +
		fromFiles +
		combineGroup +
		whereRelNull +
		magazine +
		groupbyRel +
		filterCountAndYear +
		deletedatIsNull +
		orderMinYearRel
}

// ScenerSQL is the SQL query for getting sceners.
func ScenerSQL(name string) string {
	n := strings.ToUpper(releaser.Humanize(name))
	exact := fmt.Sprintf("(upper(credit_text) = '%s')"+
		" OR (upper(credit_program) = '%s')"+
		" OR (upper(credit_illustration) = '%s')"+
		" OR (upper(credit_audio) = '%s')", n, n, n, n)
	first := fmt.Sprintf("(upper(credit_text) LIKE '%s,%%')"+
		" OR (upper(credit_program) LIKE '%s,%%')"+
		" OR (upper(credit_illustration) LIKE '%s,%%')"+
		" OR (upper(credit_audio) LIKE '%s,%%')", n, n, n, n)
	middle := fmt.Sprintf("(upper(credit_text) LIKE '%%,%s,%%')"+
		" OR (upper(credit_program) LIKE '%%,%s,%%')"+
		" OR (upper(credit_illustration) LIKE '%%,%s,%%')"+
		" OR (upper(credit_audio) LIKE '%%,%s,%%')", n, n, n, n)
	last := fmt.Sprintf("(upper(credit_text) LIKE '%%,%s')"+
		" OR (upper(credit_program) LIKE '%%,%s')"+
		" OR (upper(credit_illustration) LIKE '%%,%s')"+
		" OR (upper(credit_audio) LIKE '%%,%s')", n, n, n, n)
	return fmt.Sprintf("(%s) OR (%s) OR (%s) OR (%s)", exact, first, middle, last)
}

func Releasers() SQL {
	return "SELECT DISTINCT releaser " +
		fromFiles +
		combineGroup +
		whereRelNull +
		"GROUP BY releaser " +
		"ORDER BY releaser ASC"
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

// SimilarToReleaser selects a list of distinct releasers or groups,
// like the query strings and ordered by the file count.
func SimilarToReleaser(like ...string) SQL {
	query := like
	for i, val := range query {
		query[i] = strings.ToUpper(strings.TrimSpace(val))
	}
	return "SELECT * FROM (" + releaserSEL + releaserBy +
		SQL(fmt.Sprintf(") sub WHERE sub.releaser SIMILAR TO '%%(%s)%%'", strings.Join(query, "|"))) +
		" ORDER BY sub.count_sum DESC"
}

// SimilarToMagazine selects a list of distinct magazine titles,
// like the query strings and ordered by the file count.
func SimilarToMagazine(like ...string) SQL {
	query := like
	for i, val := range query {
		query[i] = strings.ToUpper(strings.TrimSpace(val))
	}
	return "SELECT * FROM (" + releaserSEL + magazine + releaserBy +
		SQL(fmt.Sprintf(") sub WHERE sub.releaser SIMILAR TO '%%(%s)%%'", strings.Join(query, "|"))) +
		" ORDER BY sub.count_sum DESC"
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
		fmt.Sprintf("CROSS JOIN LATERAL (values%s) AS T(scener) ", r) +
		"WHERE NULLIF(scener, '') IS NOT NULL " + // exclude empty scener names
		"GROUP BY scener, files.deletedat " + // group the results by the scener name and deleteat
		// filter the results by the file count and only include releasers with public records
		"HAVING (COUNT(files.filename) > 0) AND files.deletedat IS NULL " +
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

// SumSection is an SQL statement to sum the filesizes of records matching the section.
func SumSection() SQL {
	return "SELECT SUM(files.filesize) FROM files WHERE section = $1"
}

// SumGroup is an SQL statement to sum the filesizes of records matching the group.
func SumGroup() SQL {
	return "SELECT SUM(filesize) as size_total FROM files WHERE group_brand_for = $1"
}

// SumPlatform is an SQL statement to sum the filesizes of records matching the platform.
func SumPlatform() SQL {
	return "SELECT sum(filesize) FROM files WHERE platform = $1"
}

// SetUpper is an SQL statement to update a column with uppercase values.
func SetUpper(column string) string {
	return "UPDATE files " +
		fmt.Sprintf("SET %s = UPPER(%s);", column, column)
}

// SetFilesize0 is an SQL statement to update filesize column NULLs with 0 values.
// This is a fix for the error: failed to bind pointers to obj: sql:
// Scan error on column index 2, name "size_total": converting NULL to int is unsupported.
func SetFilesize0() string {
	return "UPDATE files SET filesize = 0 WHERE filesize IS NULL;"
}
