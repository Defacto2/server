package postgres

// Package file sql.go contains custom SQL statements that cannot be created using the SQLBoiler tool.

import (
	"fmt"
	"strconv"
	"strings"
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
		return err
	}
	rows, err := conn.Query(Ver)
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	defer conn.Close()
	for rows.Next() {
		if err := rows.Scan(v); err != nil {
			return err
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

// releaserSEL is a partial SQL statement to select the releasers name, file count and filesize sum.
const releaserSEL SQL = "SELECT DISTINCT releaser, " + // select distinct releaser names
	"COUNT(files.filename) AS count_sum, " + // count the number of files per releaser
	"SUM(files.filesize) AS size_total " + // sum the filesize of files per releaser
	"FROM files " +
	// combine the group_brand_for and group_brand_by columns as releasers
	"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
	"WHERE NULLIF(releaser, '') IS NOT NULL " // exclude empty releaser names

// releaserBy is a partial SQL statement to group the results by the releaser name.
const releaserBy SQL = "GROUP BY releaser ORDER BY releaser ASC"

// ReleasersAlphabetical selects a list of distinct releasers or groups,
// excluding BBS and FTP sites.
func ReleasersAlphabetical() SQL {
	return releaserSEL +
		"AND releaser !~ 'BBS\\M' AND releaser !~ 'FTP\\M' " + // exclude BBS and FTP sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		"HAVING (COUNT(files.filename) > 0) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY releaser ASC" // order the list by the oldest year and the releaser name
}

// BBSsAlphabetical selects a list of distinct BBS names.
func BBSsAlphabetical() SQL {
	return releaserSEL +
		"AND releaser ~ 'BBS\\M' " + // require BBS sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		"HAVING (COUNT(files.filename) > 0) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY releaser ASC" // order the list by the oldest year and the releaser name
}

// FTPsAlphabetical selects a list of distinct FTP site names.
func FTPsAlphabetical() SQL {
	return releaserSEL +
		"AND releaser ~ 'FTP\\M' " + // require FTP sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		"HAVING (COUNT(files.filename) > 0) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY releaser ASC" // order the list by the oldest year and the releaser name
}

// MagazinesAlphabetical selects a list of distinct magazine titles.
func MagazinesAlphabetical() SQL {
	return releaserSEL +
		"AND section = 'magazine' " + // require magazines
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		"HAVING (COUNT(files.filename) > 0) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY releaser ASC" // order the list by the oldest year and the releaser name
}

// ReleasersProlific selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the file count.
func ReleasersProlific() SQL {
	return releaserSEL +
		"AND releaser !~ 'BBS\\M' AND releaser !~ 'FTP\\M' " + // exclude BBS and FTP sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		"HAVING (COUNT(files.filename) > 0) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY count_sum DESC, releaser ASC" // order the list by the oldest year and the releaser name
}

// BBSsProlific selects a list of distinct releasers or groups,
// only showing BBS sites and ordered by the file count.
func BBSsProlific() SQL {
	return releaserSEL +
		"AND releaser ~ 'BBS\\M' " + // require BBS sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		"HAVING (COUNT(files.filename) > 0) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY count_sum DESC, releaser ASC" // order the list by the oldest year and the releaser name
}

// ReleasersOldest selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the oldest year.
func ReleasersOldest() SQL {
	return "SELECT DISTINCT releaser, " + // select distinct releaser names
		"COUNT(files.filename) AS count_sum, " + // count the number of files per releaser
		"SUM(files.filesize) AS size_total, " + // sum the filesize of files per releaser
		"MIN(files.date_issued_year) AS min_year " + // select the oldest year per releaser
		"FROM files " +
		// combine the group_brand_for and group_brand_by columns as releasers
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
		"WHERE NULLIF(releaser, '') IS NOT NULL " + // exclude empty releaser names
		"AND releaser !~ 'BBS\\M' AND releaser !~ 'FTP\\M' " + // exclude BBS and FTP sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		// filter the results by the file count and the oldest year,
		// this is to exclude releasers with less than 1 file or an unknown release year
		"HAVING (COUNT(files.filename) > 0) AND (MIN(files.date_issued_year) > 1970) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY min_year ASC, releaser ASC" // order the list by the oldest year and the releaser name
}

// BBSsOldest selects a list of distinct releasers or groups,
// only showing BBS sites and ordered by the file count.
func BBSsOldest() SQL {
	return "SELECT DISTINCT releaser, " + // select distinct releaser names
		"COUNT(files.filename) AS count_sum, " + // count the number of files per releaser
		"SUM(files.filesize) AS size_total, " + // sum the filesize of files per releaser
		"MIN(files.date_issued_year) AS min_year " + // select the oldest year per releaser
		"FROM files " +
		// combine the group_brand_for and group_brand_by columns as releasers
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
		"WHERE NULLIF(releaser, '') IS NOT NULL " + // exclude empty releaser names
		"AND releaser ~ 'BBS\\M' " + // require BBS sites
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		// filter the results by the file count and the oldest year,
		// this is to exclude releasers with less than 1 file or an unknown release year
		"HAVING (COUNT(files.filename) > 0) AND (MIN(files.date_issued_year) > 1970) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY min_year ASC, releaser ASC" // order the list by the oldest year and the releaser name
}

func MagazinesOldest() SQL {
	return "SELECT DISTINCT releaser, " + // select distinct releaser names
		"COUNT(files.filename) AS count_sum, " + // count the number of files per releaser
		"SUM(files.filesize) AS size_total, " + // sum the filesize of files per releaser
		"MIN(files.date_issued_year) AS min_year " + // select the oldest year per releaser
		"FROM files " +
		// combine the group_brand_for and group_brand_by columns as releasers
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
		"WHERE NULLIF(releaser, '') IS NOT NULL " + // exclude empty releaser names
		"AND section = 'magazine' " + // require magazines
		"GROUP BY releaser, files.deletedat " + // group the results by the releaser name and deleteat
		// filter the results by the file count and the oldest year,
		// this is to exclude releasers with less than 1 file or an unknown release year
		"HAVING (COUNT(files.filename) > 0) AND (MIN(files.date_issued_year) > 1970) " +
		"AND files.deletedat IS NULL " + // only include releasers with public records
		"ORDER BY min_year ASC, releaser ASC" // order the list by the oldest year and the releaser name
}

// ReleaserSimilarTo selects a list of distinct releasers or groups,
// like the query strings and ordered by the file count.
func ReleaserSimilarTo(like ...string) SQL {
	query := like
	for i, val := range query {
		query[i] = strings.ToUpper(strings.TrimSpace(val))
	}
	return "SELECT * FROM (" + releaserSEL + releaserBy +
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
		"FROM files " +
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
