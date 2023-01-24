// Package models contain the custom queries for the database that are not available using the ORM,
// as well as methods to interact with the query data.
package models

// File is an object representing the database table.
// type File struct { // Primary key
// 	ID int64 `boil:"id" json:"id" toml:"id" yaml:"id"`
// }

// https://github.com/volatiletech/sqlboiler#constants

type Count int // Count is the number of found files.

// Scener contains the usable data for a group or person.
type Scener struct {
	URI   string // URI slug for the scener.
	Name  string // Name to display.
	Count int    // Count the records associated with the scene.
}

// Sceners is a collection of sceners.
type Sceners map[string]Scener

// Counts caches the number of found files fetched from SQL queries.
var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
}

const (
	Art  int = iota // Art are digital + pixel art files.
	Doc             // Doc are document + text art files.
	Soft            // Soft are software files.

	groupFor   = "group_brand_for = ?"
	section    = "section = ?"
	notSection = "section != ?"
	platform   = "platform = ?"
)
