package html3

import (
	"fmt"
)

// Navigate handles offset and record limit pagination.
type Navigate struct {
	Current  string // Current name of the active record query.
	QueryStr string // QueryStr to append to all pagination links.
	Limit    int    // Limit the number of records to return per query.
	Link1    int    // Link1 of the dynamic pagination.
	Link2    int    // Link2 of the dynamic pagination.
	Link3    int    // Link3 of the dynamic pagination.
	Page     int    // Page number of the current record query.
	PagePrev int    // PagePrev is the page number to the previous record query.
	PageNext int    // PageNext is the page number to the next record query.
	PageMax  int    // PageMax is the maximum and last page number of the record query.
}

// Navi returns a populated Navigate struct for pagination.
func Navi(limit, page int, maxPage uint, current, qs string) Navigate {
	return Navigate{
		Current:  current,
		Limit:    limit,
		Page:     page,
		PagePrev: previous(page),
		PageNext: next(page, maxPage),
		PageMax:  int(maxPage),
		QueryStr: qs,
	}
}

// qs returns a query string with a leading question mark.
func qs(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("?%s", s)
}

// previous returns the previous page number.
func previous(page int) int {
	if page == 1 {
		return 1
	}
	return page - 1
}

// next returns the next page number.
func next(page int, maxPage uint) int {
	max := int(maxPage)
	if page >= max {
		return max
	}
	return page + 1
}
