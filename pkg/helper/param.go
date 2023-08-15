package helper

// Package file param.go contains functions specific for the unique, Defacto2 ID URLs.

import (
	"math"
	"net/url"
	"path"
	"strconv"

	"github.com/bengarrett/cfw"
)

// Deobfuscate an obfuscated ID to return the primary key of the record.
// Returns a 0 if the id is not valid.
func Deobfuscate(id string) int {
	key, _ := strconv.Atoi(cfw.DeObfuscate(id))
	return key
}

// Deobfuscate an obfuscated record URL to return a record's primary key.
// A URL can point to a Defacto2 record download or detail page.
// Returns a 0 if the URL is not valid.
func DeobfuscateURL(rawURL string) int {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	return Deobfuscate(path.Base(u.Path))
}

// Obfuscates the primary key of a record as a string that is used as a URL param or path.
func Obfuscate(key int64) string {
	return cfw.Obfuscate(strconv.Itoa(int(key)))
}

// PageCount returns the maximum pages possible for the sum of records with a record limit per-page.
func PageCount(sum, limit int) uint {
	if sum <= 0 || limit <= 0 {
		return 0
	}
	x := math.Ceil(float64(sum) / float64(limit))
	return uint(x)
}

// Pagination returns the page numbers to display in the pagination bar.
func Pagination(page, sum, limit int) []int {
	if page <= 0 || sum <= 0 || limit <= 0 {
		return nil
	}
	const firstPage = 1
	lastPage := int(PageCount(sum, limit))
	pages := make([]int, 0, 9)
	pages = append(pages, firstPage)
	pages = append(pages, lastPage)

	// pages = append(pages, 1)
	// p := int(PageCount(sum, limit))
	// for i := 1; i < 9; i++ {
	// 	if page-i > 1 {
	// 		pages = append(pages, page-i)
	// 		fmt.Fprintln(os.Stderr, ">>", i, page-i)
	// 	}
	// 	if page+i < p {
	// 		pages = append(pages, page+i)
	// 		fmt.Fprintln(os.Stderr, "<<", i, page+1)
	// 	}
	// }
	// pages = append(pages, p)
	return pages
}
