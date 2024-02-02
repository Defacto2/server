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
		return fmt.Sprintf("and using %s", strings.Join(x[0:2], " "))
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
// The distinct on clause is used in PostreSQL to create a non-case sensitive list of releasers.
// It requires a matching order by upper clause to be valid.
// The cross join lateral clause is used to create a distinct list of releasers from
// the group_brand_for and group_brand_by columns.
//
// Note these SQLs may cause inconsistent results when used with the count_sum and size_total columns.
// This is because there SelectRels and SelectRelsPros excludes some files from the count and sum.
const releaserSEL SQL = "SELECT DISTINCT releaser, " +
	"COUNT(files.filename) AS count_sum, " +
	"SUM(files.filesize) AS size_total " +
	"FROM files " +
	"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
	"WHERE NULLIF(releaser, '') IS NOT NULL "

// releaserBy is a partial SQL statement to group the results by the releaser name.
const releaserBy SQL = "GROUP BY releaser " +
	"ORDER BY releaser ASC"

const magazineSEL SQL = "SELECT DISTINCT releaser, " +
	"COUNT(files.filename) AS count_sum, " +
	"SUM(files.filesize) AS size_total, " +
	"MIN(files.date_issued_year) AS min_year " +
	"FROM files " +
	"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
	"WHERE NULLIF(releaser, '') IS NOT NULL " +
	"AND section = 'magazine' " +
	"GROUP BY releaser " +
	"ORDER BY min_year ASC, releaser ASC"

// Roles returns all of the sceners reguardless of the attribution.
func Roles() Role {
	s := strings.Join([]string{string(Writer), string(Artist), string(Coder), string(Musician)}, ",")
	return Role(s)
}

func (r Role) Select() SQL {
	s := "SELECT DISTINCT ON(upper(scener)) scener " +
		"FROM files " +
		fmt.Sprintf("CROSS JOIN LATERAL (values%s) AS T(scener) ", r) +
		"WHERE NULLIF(scener, '') IS NOT NULL " +
		"GROUP BY scener " +
		"ORDER BY upper(scener) ASC"
	return SQL(s)
}

// DistScener selects a list of distinct sceners.
func DistScener() SQL {
	return Roles().Select()
}

// DistWriter selects a list of distinct writers.
func DistWriter() SQL {
	return Writer.Select()
}

// DistArtist selects a list of distinct artists.
func DistArtist() SQL {
	return Artist.Select()
}

// DistCoder selects a list of distinct coders.
func DistCoder() SQL {
	return Coder.Select()
}

// DistMusician selects a list of distinct musicians.
func DistMusician() SQL {
	return Musician.Select()
}

// DistReleaser selects a list of distinct releasers or groups,
// excluding BBS and FTP sites.
func DistReleaser() SQL {
	return releaserSEL +
		"AND releaser !~ 'BBS\\M' " +
		"AND releaser !~ 'FTP\\M' " +
		releaserBy
}

// SelectRelsPros selects a list of distinct releasers or groups,
// excluding BBS and FTP sites and ordered by the file count.
func DistReleaserSummed() SQL {
	return "SELECT * FROM (" +
		releaserSEL +
		"AND releaser !~ 'BBS\\M' " +
		"AND releaser !~ 'FTP\\M' " +
		releaserBy +
		") sub WHERE sub.count_sum > 2 ORDER BY sub.count_sum DESC"
}

// DistMagazine selects a list of distinct magazine titles.
func DistMagazine() SQL {
	return releaserSEL + "AND section = 'magazine'" + releaserBy
}

func DistMagazineByYear() SQL {
	return magazineSEL
}

// DistBBS selects a list of distinct BBS names.
func DistBBS() SQL {
	return releaserSEL + "AND releaser ~ 'BBS\\M' " + releaserBy
}

// DistBBSSummed selects a list of distinct BBS names ordered by the file count.
func DistBBSSummed() SQL {
	return "SELECT * FROM (" +
		releaserSEL +
		"AND releaser ~ 'BBS\\M' " +
		releaserBy +
		") sub WHERE sub.count_sum > 2 ORDER BY sub.count_sum DESC"
}

// DistFTP selects a list of distinct FTP site names.
func DistFTP() SQL {
	return releaserSEL + "AND releaser ~ 'FTP\\M' " + releaserBy
}

// SumReleaser is an SQL statement to total the file count and filesize sum of releasers,
// as well as the minimum, oldest and maximum, newest year values.
// The where parameter is used to filter the releasers by section, either all, magazine, bbs or ftp.
func SumReleaser(where string) SQL {
	s := "SELECT COUNT(files.id) AS count_total, " +
		"SUM(files.filesize) AS size_total, " +
		"MIN(files.date_issued_year) AS min_year, " +
		"MAX(files.date_issued_year) AS max_year " +
		"FROM files "
	switch where {
	case "magazine":
		s += "WHERE files.section = 'magazine'"
	case "bbs":
		s += "WHERE files.group_brand_for ~ 'BBS\\M' " +
			"OR files.group_brand_by ~ 'BBS\\M'"
	case "ftp":
		s += "WHERE files.group_brand_for ~ 'FTP\\M' " +
			"OR files.group_brand_by ~ 'FTP\\M'"
	default:
		return ""
	}
	return SQL(strings.TrimSpace(s))
}

// SumBBS is an SQL statement to total the file count and filesize sum of BBS sites.
func SumBBS() SQL {
	return SumReleaser("bbs")
}

// SumFTP is an SQL statement to total the file count and filesize sum of FTP sites.
func SumFTP() SQL {
	return SumReleaser("ftp")
}

// SumMag is an SQL statement to total the file count and filesize sum of magazine titles.
func SumMag() SQL {
	return SumReleaser("magazine")
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
