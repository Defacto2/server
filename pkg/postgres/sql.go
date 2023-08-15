package postgres

// Package file sql.go contains custom SQL statements that cannot be created using the SQLBoiler tool.

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Counter is a partial SQL statement to count the number of records.
	Counter = "COUNT(*) AS counter"
	// MinYear is a partial SQL statement to select the minimum year value.
	MinYear = "MIN(date_issued_year) AS min_year"
	// MaxYear is a partial SQL statement to select the maximum year value.
	MaxYear = "MAX(date_issued_year) AS max_year"
	// SumSize is a partial SQL statement to sum the filesize values of multiple records.
	SumSize = "SUM(filesize) AS size_sum"
	// Ver is a SQL statement to select the version of the PostgreSQL database server in use.
	Ver = "SELECT version();"

	Totals = "COUNT(*) AS count_total, SUM(filesize) AS size_total"
	Years  = "MIN(date_issued_year) AS min_year, MAX(date_issued_year) AS max_year"
)

func Statistics() []string {
	return []string{Totals, Years}
}

// Columns returns a list of column selections.
func Columns() []string {
	return []string{SumSize, Counter, MinYear, MaxYear}
}

// Stat returns the SumSize and Counter column selections.
func Stat() []string {
	return []string{SumSize, Counter}
}

type Version string // Version of the PostgreSQL database server in use.

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

type SQL string // SQL is a raw query statement for PostgreSQL.

// releaserSEL is a partial SQL statement to select the releasers name, file count and filesize sum.
// The distinct on clause is used in PostreSQL to create a non-case sensitive list of releasers.
// It requires a matching order by upper clause to be valid.
// The cross join lateral clause is used to create a distinct list of releasers from
// the group_brand_for and group_brand_by columns.
const releaserSEL SQL = "SELECT DISTINCT ON(upper(releaser)) releaser, " +
	"COUNT(files.filename) AS count_sum, " +
	"SUM(files.filesize) AS size_sum " +
	"FROM files " +
	"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
	"WHERE NULLIF(releaser, '') IS NOT NULL "

const releaserBy SQL = "GROUP BY releaser " +
	"ORDER BY upper(releaser) ASC"

// SelectRels selects a list of distinct releasers or groups,
// excluding magazine titles, BBS and FTP sites.
func SelectRels() SQL {
	return releaserSEL +
		"AND section != 'magazine' " +
		"AND releaser !~ 'BBS\\M' " +
		"AND releaser !~ 'FTP\\M' " +
		releaserBy
}

func SelectRelPros() SQL {
	return "SELECT * FROM (" +
		releaserSEL +
		"AND section != 'magazine' " +
		"AND releaser !~ 'BBS\\M' " +
		"AND releaser !~ 'FTP\\M' " +
		releaserBy +
		") sub ORDER BY sub.count_sum DESC"
}

// SelectMags selects a list of distinct magazine titles.
func SelectMag() SQL {
	return releaserSEL + "AND section = 'magazine'" + releaserBy
}

// SelectBBS selects a list of distinct BBS names.
func SelectBBS() SQL {
	return releaserSEL + "AND releaser ~ 'BBS\\M' " + releaserBy
}

func SelectBBSPros() SQL {
	return "SELECT * FROM (" +
		releaserSEL +
		"AND releaser ~ 'BBS\\M' " +
		releaserBy +
		") sub ORDER BY sub.count_sum DESC"
}

// SelectFTP selects a list of distinct FTP site names.
func SelectFTP() SQL {
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
	return "SELECT SUM(filesize) as size_sum FROM files WHERE group_brand_for = $1"
}

// SumPlatform is an SQL statement to sum the filesizes of records matching the platform.
func SumPlatform() SQL {
	return "SELECT sum(filesize) FROM files WHERE platform = $1"
}
