package app

// Package file uri.go contains the URI type, strings and methods.

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
	deletions
	demoscene
	drama
	forApproval
	ftp
	hack
	howTo
	htm
	java
	jobAdvert
	imageFile
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
	unwanted
	video
	windows
	windowsPack
)

func (u URI) String() string { //nolint:funlen
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
		"deletions",
		"demoscene",
		"drama",
		"for-approval",
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
		"unwanted",
		"video",
		"windows",
		"windows-pack",
	}[u]
}

// Match path to a URI type or return -1 if not found.
func Match(path string) URI {
	for val := range int(windowsPack) {
		i := val + 1
		if URI(i).String() == path {
			return URI(i)
		}
	}
	return -1
}

// Valid returns true if path is a valid URI for the list of files.
func Valid(path string) bool {
	for val := range int(windowsPack) {
		i := val + 1
		if URI(i).String() == path {
			return true
		}
	}
	return false
}
