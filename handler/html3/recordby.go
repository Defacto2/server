package html3

// Package file recordby.go contains the record grouping functions.

const (
	title   = "Index of " + Prefix
	textAll = "list every file or release hosted on the website"
	textArt = "hi-res, raster and pixel images"
	textDoc = "documents using any media format, including text files, ASCII, and ANSI text art"
	textSof = "applications and programs for any platform"
	firefox = "Welcome to the Firefox v2, 2006 era, Defacto2 website, " +
		"which is friendly for legacy operating systems, including Windows 9x, NT-4, and OS-X 10.2."
)

// RecordsBy are the record groupings.
type RecordsBy int

const (
	Everything RecordsBy = iota // Everything displays all records from the file table.
	BySection                   // BySection groups records by the section file table column.
	ByPlatform                  // BySection groups records by the platform file table column.
	ByGroup                     // ByGroup groups the records by the distinct, group_brand_for file table column.
	AsArt                       // AsArt group records as art.
	AsDocument                  // AsDocument group records as documents.
	AsSoftware                  // AsSoftware group records as software.
)

// Parent returns the parent route for the current route.
func (t RecordsBy) Parent() string {
	const l = 7
	if t >= l {
		return ""
	}
	const blank = ""
	return [l]string{
		blank,
		"categories",
		"platforms",
		"groups",
		blank,
		blank,
		blank,
	}[t]
}

// String RecordsBy are the record groupings.
func (t RecordsBy) String() string {
	const l = 7
	if t >= l {
		return ""
	}
	return [l]string{
		"html3_all",
		"html3_category",
		"html3_platform",
		"html3_group",
		"html3_art",
		"html3_documents",
		"html3_software",
	}[t]
}
