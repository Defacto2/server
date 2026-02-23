// Package extensions provides a list of file extensions used by some functions in app.
package extensions

const (
	avif = ".avif"
	fzip = ".zip"
	gif  = ".gif"
	jpeg = ".jpeg"
	jpg  = ".jpg"
	png  = ".png"
	webp = ".webp"
)

var (
	archiveExts  = []string{fzip, ".rar", ".7z", ".tar", ".lha", ".lzh", ".arc", ".arj", ".ace"}
	documentExts = []string{
		".txt", ".nfo", ".diz", ".asc", ".lit", ".rtf", ".doc", ".docx",
		".pdf", ".unp", ".htm", ".html", ".xml", ".json", ".csv",
	}
	imageExts = []string{avif, gif, jpg, jpeg, ".jfif", png, ".svg", webp, ".bmp", ".ico"}
	mediaExts = []string{".mpeg", ".mp1", ".mp2", ".mp3", ".mp4", ".ogg", ".webm", ".wmv"}
)

// Archive returns a list of archive file extensions supported by this web application.
func Archive() []string {
	return archiveExts
}

// Document returns a list of document file extensions that can be read as text in the browser.
func Document() []string {
	return documentExts
}

// Image returns a list of image file extensions that can be displayed in the browser.
func Image() []string {
	return imageExts
}

// Media returns a list of [media file extensions] that can be played in the browser.
//
// [media file extensions]: https://developer.mozilla.org/en-US/docs/Web/Media/Formats
func Media() []string {
	return mediaExts
}
