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
)

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

const releaserSEL = "SELECT DISTINCT group_brand, " +
	"COUNT(group_brand) AS count, " +
	"SUM(files.filesize) AS size_sum " +
	"FROM files " +
	"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) " +
	"WHERE NULLIF(group_brand, '') IS NOT NULL "

const releaserBy = "GROUP BY group_brand " +
	"ORDER BY group_brand ASC"

// SelectRelr selects a list of distinct releasers or groups.
func SelectRelr() SQL {
	return releaserSEL + releaserBy
}

// SelectMags selects a list of distinct magazine titles.
func SelectMag() SQL {
	return releaserSEL + "AND section = 'magazine'" + releaserBy
}

// SelectBBS selects a list of distinct BBS names.
func SelectBBS() SQL {
	return releaserSEL + "AND group_brand ~ 'BBS\\M'" + releaserBy
}

// SelectFTP selects a list of distinct FTP site names.
func SelectFTP() SQL {
	return releaserSEL + "AND group_brand ~ 'FTP\\M'" + releaserBy
}

// StatRelr is an SQL statement to select all the unique groups.
func StatRelr() SQL {
	return "SELECT DISTINCT group_brand FROM files " +
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) " +
		"WHERE NULLIF(group_brand, '') IS NOT NULL " + // handle empty and null values
		"GROUP BY group_brand"
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
