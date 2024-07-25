package readme

import (
	"cmp"
	"path/filepath"
	"slices"
	"strings"
)

// ReadmeSuggest returns a suggested readme file name for the record.
// It prioritizes the filename and group name with a priority extension,
// such as ".nfo", ".txt", etc. If no priority extension is found,
// it will return the first textfile in the content list.
//
// The filename should be the name of the file archive artifact.
// The group should be a name or common abbreviation of the group that
// released the artifact. The content should be a list of files contained
// in the artifact.
//
// This is a port of the CFML function, variables.findTextfile found in File.cfc.
func Suggest(filename, group string, content ...string) string {
	finds := Readmes(content...)
	if len(finds) == 1 {
		return finds[0]
	}
	finds = SortContent(finds...)

	// match either the filename or the group name with a priority extension
	// e.g. .nfo, .txt, .unp, .doc
	base := filepath.Base(filename)
	for _, ext := range priority() {
		for _, name := range finds {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	const matchFileID = "file_id.diz"
	for _, name := range finds {
		if strings.EqualFold(matchFileID, name) {
			return name
		}
	}
	// match either the filename or the group name with a candidate extension
	for _, ext := range candidate() {
		for _, name := range finds {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match any finds that use a priority extension
	for _, name := range finds {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			return name
		}
	}
	// match the first file in the list
	for _, name := range finds {
		return name
	}
	return ""
}

// Readmes returns a list of readme text files found in the file archive.
func Readmes(content ...string) []string {
	finds := []string{}
	skip := []string{"scene.org", "scene.org.txt"}
	for _, name := range content {
		if name == "" {
			continue
		}
		s := strings.ToLower(name)
		if slices.Contains(skip, s) {
			continue
		}
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			finds = append(finds, name)
			continue
		}
		if slices.Contains(candidate(), ext) {
			finds = append(finds, name)
		}
	}
	return finds
}

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return []string{".nfo", ".txt", ".unp", ".doc"}
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
}

// SortContent sorts the content list by the number of slashes in each string.
// It prioritizes strings with fewer slashes (i.e., closer to the root).
// If the number of slashes is the same, it sorts alphabetically.
func SortContent(content ...string) []string {
	const windowsPath = "\\"
	const pathSeparator = "/"
	slices.SortFunc(content, func(a, b string) int {
		a = strings.ReplaceAll(a, windowsPath, pathSeparator)
		b = strings.ReplaceAll(b, windowsPath, pathSeparator)
		aCount := strings.Count(a, pathSeparator)
		bCount := strings.Count(b, pathSeparator)
		if aCount != bCount {
			return aCount - bCount
		}
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return content
}
