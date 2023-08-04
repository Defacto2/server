package app

// URI is a type for the files URI path.
type URI int

const (
	root URI = iota
	advert
	announcement
	ansi
	ansiBBS
	ansiBrand
	ansiFTP
	ansiPack
	ansiNfo
	ansiTool
	bbs
	bbstro
	bbsImage
	bbsText
	database
	demoscene
	drama
	ftp
	hack
	howTo
	html
	java
	jobAdvert
	image
	imagePack
	intro
	introMsdos
	introWindows
	installer
	linux
	magazine
	macos
	msdos
	msdosPack
	music
	newest
	newsArticle
	newUpdates
	newUploads
	nfo
	nfoTool
	oldest
	pdf
	proof
	restrict
	script
	standards
	takedown
	text
	textAmiga
	textApple2
	textAtariST
	textPack
	tool
	trialCrackme
	video
	windows
	windowsPack
)

func (u URI) String() string {
	return [...]string{
		"",
		"advert",
		"announcement",
		"ansi",
		"ansi-bbs",
		"ansi-brand",
		"ansi-ftp",
		"ansi-pack",
		"ansi-nfo",
		"ansi-tool",
		"bbs",
		"bbstro",
		"bbs-image",
		"bbs-text",
		"database",
		"demoscene",
		"drama",
		"ftp",
		"hack",
		"how-to",
		"html",
		"java",
		"job-advert",
		"image",
		"image-pack",
		"intro",
		"intro-msdos",
		"intro-windows",
		"installer",
		"linux",
		"magazine",
		"macos",
		"msdos",
		"msdos-pack",
		"music",
		"newest",
		"news-article",
		"new-updates",
		"new-uploads",
		"nfo",
		"nfo-tool",
		"oldest",
		"pdf",
		"proof",
		"restrict",
		"script",
		"standards",
		"takedown",
		"text",
		"text-amiga",
		"text-apple2",
		"text-atari-st",
		"text-pack",
		"tool",
		"trial-crackme",
		"video",
		"windows",
		"windows-pack",
	}[u]
}

// IsURI checks if the string is a valid files URI path.
func IsURI(s string) bool {
	// range to 57
	for i := 1; i <= int(windowsPack); i++ {
		if URI(i).String() == s {
			return true
		}
	}
	return false
}

// Match string to URI type or return -1 if not found.
func Match(s string) URI {
	// range to 57
	for i := 1; i <= int(windowsPack); i++ {
		if URI(i).String() == s {
			return URI(i)
		}
	}
	return -1
}
