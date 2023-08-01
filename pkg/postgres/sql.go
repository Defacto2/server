package postgres

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
)

type Version string // Version of the PostgreSQL database server in use.

// Query the database version.
func (v *Version) Query() error {
	conn, err := ConnectDB()
	if err != nil {
		return err
	}
	rows, err := conn.Query("SELECT version();")
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
		return fmt.Sprintf("using %s", strings.Join(x[0:2], " "))
	}
	return s
}

// SQLGroupStat is an SQL statement to select all the unique groups.
func SQLGroupStat() string {
	return "SELECT DISTINCT group_brand FROM files " +
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) " +
		"WHERE NULLIF(group_brand, '') IS NOT NULL " + // handle empty and null values
		"GROUP BY group_brand"
}

// SQLGroupAll is an SQL statement to collect statistics for each of the unique groups.
func SQLGroupAll() string {
	return "SELECT DISTINCT group_brand, " +
		"COUNT(group_brand) AS count, " +
		"SUM(files.filesize) AS size_sum " +
		"FROM files " +
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) " +
		"WHERE NULLIF(group_brand, '') IS NOT NULL " + // handle empty and null values
		"GROUP BY group_brand " +
		"ORDER BY group_brand ASC"
}

// SQLSumSection is an SQL statement to sum the filesizes of records matching the section.
func SQLSumSection() string {
	return "SELECT SUM(files.filesize) FROM files WHERE section = $1"
}

// SQLSumGroup is an SQL statement to sum the filesizes of records matching the group.
func SQLSumGroup() string {
	return "SELECT SUM(filesize) as size_sum FROM files WHERE group_brand_for = $1"
}

// SQLSumPlatform is an SQL statement to sum the filesizes of records matching the platform.
func SQLSumPlatform() string {
	return "SELECT sum(filesize) FROM files WHERE platform = $1"
}
