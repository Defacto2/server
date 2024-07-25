package exts

const (
	avif      = ".avif"
	fzip      = ".zip"
	gif       = ".gif"
	jpeg      = ".jpeg"
	jpg       = ".jpg"
	png       = ".png"
	textamiga = "textamiga"
	webp      = ".webp"
)

// archives returns a list of archive file extensions supported by this web application.
func Archives() []string {
	return []string{fzip, ".rar", ".7z", ".tar", ".lha", ".lzh", ".arc", ".arj", ".ace", ".tar"}
}

// documents returns a list of document file extensions that can be read as text in the browser.
func Documents() []string {
	return []string{
		".txt", ".nfo", ".diz", ".asc", ".lit", ".rtf", ".doc", ".docx",
		".pdf", ".unp", ".htm", ".html", ".xml", ".json", ".csv",
	}
}

// images returns a list of image file extensions that can be displayed in the browser.
func Images() []string {
	return []string{".avif", gif, jpg, jpeg, ".jfif", png, ".svg", webp, ".bmp", ".ico"}
}

// media returns a list of [media file extensions] that can be played in the browser.
//
// [media file extensions]: https://developer.mozilla.org/en-US/docs/Web/Media/Formats
func Media() []string {
	return []string{".mpeg", ".mp1", ".mp2", ".mp3", ".mp4", ".ogg", ".wmv"}
}
