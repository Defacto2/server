//nolint:gochecknoglobals
package readme

import (
	"cmp"
	"path/filepath"
	"slices"
	"strings"
)

// Suggest returns a suggested readme file name for the record.
// It prioritizes the filename and group name with a priority extension,
// such as ".nfo", ".txt", etc. If no priority extension is found,
// it will return the first text file in the content list.
//
//   - The name string should be the original filename of the archive.
//   - The group should be a known name or common abbreviation,
//   - The content should be a list of the archive entries.
func Suggest(name, group, entries string) string {
	finds := SortList(true, entries)
	if len(finds) == 0 {
		return ""
	}
	if len(finds) == 1 {
		return finds[0]
	}

	base := filepath.Base(name)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}

	extensions := append(priority(), candidate()...)
	for _, ext := range extensions {
		groupMatch := group + ext
		baseMatch := base + ext

		for _, name := range finds {
			if strings.EqualFold(groupMatch, name) || strings.EqualFold(baseMatch, name) {
				return name
			}
		}
	}

	priorities := priority()
	for _, name := range finds {
		ext := strings.ToLower(filepath.Ext(name))
		if slices.Contains(priorities, ext) {
			return name
		}
	}

	return finds[0]
}

// SortList returns a sorted list of possible readme text files found in the file archive.
// The first result is the closes filename to root that has a priority
// filename extension such as ".nfo", then ordered alphabetically.
//
// When compact is true all filenames using extensions that are not known textfiles,
// are removed from the slice.
//
// To save memory, content is not split into a slice until we need to handle it.
func SortList(compact bool, entries string) []string {
	if entries == "" {
		return nil
	}

	lists := strings.Split(entries, "\n")

	slices.SortFunc(lists, func(a, b string) int {
		// used to order all other file paths alphabetically
		x := strings.ToLower(a)
		y := strings.ToLower(b)

		// used to compare the depth of sub-directories, less is better
		subsX := strings.Count(x, "/")
		subsY := strings.Count(y, "/")

		// used to order all other file extensions alphabetically
		extX := strings.ToLower(filepath.Ext(x))
		extY := strings.ToLower(filepath.Ext(y))

		rank := func(list []string, target string) int {
			if idx := slices.Index(list, target); idx != -1 {
				return idx
			}
			// Unmatched items rank lower
			const unmatched = 1000
			return unmatched
		}

		// used to compare which filename uses a candidate file extension
		prioX := rank(priority(), extX)
		prioY := rank(priority(), extY)

		// used to compare which filename uses a priority file extension
		candX := rank(candidate(), extX)
		candY := rank(candidate(), extY)

		return cmp.Or(
			cmp.Compare(subsX, subsY),
			cmp.Compare(prioX, prioY),
			cmp.Compare(candX, candY),
			strings.Compare(extX, extY),
			strings.Compare(x, y),
		)
	})

	// filter out known bad filenames, such as known website advertising injections
	paths := make([]string, 0, len(lists))

	for _, s := range lists {
		path := strings.TrimSpace(s)
		if path == "" {
			continue
		}

		base := filepath.Base(strings.ToLower(path))
		switch base {
		case "scene.org", "scene.org.txt", "we-will.sue", "www.acid.org":
			continue
		}

		ext := filepath.Ext(base)
		if compact &&
			!slices.Contains(priority(), ext) &&
			!slices.Contains(candidate(), ext) {
			continue
		}

		paths = append(paths, path)
	}
	return paths
}

var (
	prioExts = [...]string{".nfo", ".txt", ".unp", ".doc", ".displayme", ".readme"}
	candExts = [...]string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
)

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return prioExts[:]
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return candExts[:]
}
