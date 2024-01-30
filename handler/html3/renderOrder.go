package html3

import (
	"strings"

	"github.com/Defacto2/server/model/html3"
)

// Sort is the display name of column that can be used to sort and order the records.
type Sort string

const (
	NameAsc = "C=N&O=A" // Name ascending order.
	NameDes = "C=N&O=D" // Name descending order.
	PublAsc = "C=D&O=A" // Date published ascending order.
	PublDes = "C=D&O=D" // Date published descending order.
	PostAsc = "C=P&O=A" // Posted ascending order.
	PostDes = "C=P&O=D" // Posted descending order.
	SizeAsc = "C=S&O=A" // Size ascending order.
	SizeDes = "C=S&O=D" // Size descending order.
	DescAsc = "C=I&O=A" // Description ascending order.
	DescDes = "C=I&O=D" // Description descending order.

	asc  = "A" // asc is order by ascending.
	desc = "D" // desc is order by descending.
)

const (
	Name    Sort = "Name"        // Sort records by the filename.
	Publish Sort = "Publish"     // Sort records by the published year, month and day.
	Posted  Sort = "Posted"      // Sort records by the record creation dated.
	Size    Sort = "Size"        // Sort records by the file size in byte units.
	Desc    Sort = "Description" // Sort the records by the title.
)

// Sortings are the name and order of columns that the records can be ordered by.
func Sortings() map[Sort]string {
	return map[Sort]string{
		Name:    asc,
		Publish: asc,
		Posted:  asc,
		Size:    asc,
		Desc:    asc,
	}
}

// Clauses for ordering file record queries.
func Clauses(query string) html3.Order {
	switch strings.ToUpper(query) {
	case NameAsc: // Name ascending order should match the case.
		return html3.NameAsc
	case NameDes:
		return html3.NameDes
	case PublAsc:
		return html3.PublAsc
	case PublDes:
		return html3.PublDes
	case PostAsc:
		return html3.PostAsc
	case PostDes:
		return html3.PostDes
	case SizeAsc:
		return html3.SizeAsc
	case SizeDes:
		return html3.SizeDes
	case DescAsc:
		return html3.DescAsc
	case DescDes:
		return html3.DescDes
	default:
		return html3.NameAsc
	}
}

// Sorter creates the query string for the sortable columns.
// Replacing the O key value with the opposite value, either A or D.
func Sorter(query string) map[string]string {
	s := Sortings()
	switch strings.ToUpper(query) {
	case NameAsc:
		s[Name] = desc
	case NameDes:
		s[Name] = asc
	case PublAsc:
		s[Publish] = desc
	case PublDes:
		s[Publish] = asc
	case PostAsc:
		s[Posted] = desc
	case PostDes:
		s[Posted] = asc
	case SizeAsc:
		s[Size] = desc
	case SizeDes:
		s[Size] = asc
	case DescAsc:
		s[Desc] = desc
	case DescDes:
		s[Desc] = asc
	default:
		// When no query is provided, it is assumed the records have been
		// ordered with Name ASC. So set DESC for the clickable Name link.
		s[Name] = desc
	}
	// to be usable in the template, convert the map keys into strings
	fix := make(map[string]string, len(s))
	for key, value := range s {
		fix[string(key)] = value
	}
	return fix
}
