// Package tags are categories and platform metadata used to classify the served files.
//
//nolint:exhaustive
package tags

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

var (
	ErrNoDB   = errors.New("database value is nil")
	ErrNoTags = errors.New("lockable tags t is nil")
)

// NOTE: when adding a new tag make sure to update the following:
// - nameToTag size
// - LastCategory
// - LastPlatform
// - URIS{}
// - Info{}

// NOTE:on prepositions:
// Use "for" when describing subjects
//  example: an advert for a BBS
// Use "on" when targeting systems and platforms
//  example: a demo on Windows

// The dos, app funcmap handler must match the format and syntax of MS-DOS that's used here.
const msDos = "MS Dos"

// Tag is the unique ID.
type Tag int

const (
	// FirstCategory is the first tag marked as a category.
	FirstCategory Tag = Announcement
	// FirstPlatform is the first tag marked as a platform.
	FirstPlatform Tag = ANSI
	// LastCategory is the final tag marked as a category.
	LastCategory Tag = Install
	// LastPlatform is the final tag marked as a platform.
	LastPlatform Tag = Console
	// CategoryCount is the number of tags used as a category.
	CategoryCount = int(LastCategory - FirstCategory + 1)
	// PlatformCount is the number of tags used as a platform.
	PlatformCount = int(LastPlatform - FirstPlatform + 1)
)

const (
	AreaCodes Tag = iota
	Announcement
	ANSIEditor
	AppleII
	AtariST
	BBS
	Logo
	Bust
	Drama
	Rule
	Tool
	Intro
	Demo
	ForSale
	Ftp
	GameHack
	Job
	Guide
	Interview
	Mag
	News
	Nfo
	NfoTool
	Pack
	Proof
	Restrict
	Install
	ANSI
	Audio
	DataB
	DOS
	Markup
	Image
	Java
	Linux
	Mac
	PCB
	PDF
	PHP
	Text
	TextAmiga
	Video
	Windows
	Console
)

// List all the tags.
var List = [...]Tag{ //nolint:gochecknoglobals
	Announcement,
	ANSIEditor,
	AppleII,
	AtariST,
	BBS,
	Logo,
	Bust,
	Drama,
	Rule,
	Tool,
	Intro,
	Demo,
	ForSale,
	Ftp,
	GameHack,
	Job,
	Guide,
	Interview,
	Mag,
	News,
	Nfo,
	NfoTool,
	Pack,
	Proof,
	Restrict,
	Install,
	ANSI,
	Audio,
	DataB,
	DOS,
	Markup,
	Image,
	Java,
	Linux,
	Mac,
	PCB,
	PDF,
	PHP,
	Text,
	TextAmiga,
	Video,
	Windows,
	Console,
}

// Humanize returns the human readable name of the platform and section tags combined.
//
// The returned string is intended for singular artifacts and items,
// use [Humanizes] if you need plurals.
//
//   - A singular example: "a Windows intro"
//   - A plural example: "Windows intros"
func Humanize(platform, section Tag) string {
	if s := check(platform, section); s != "" {
		return s
	}

	if ppe := platform == PCB && section == Tool; ppe {
		return "a PCBoard PPE or BBS application"
	}

	if s := section.sections(platform); s != "" {
		return s
	}
	if s := platform.platforms(section); s != "" {
		return s
	}

	return genericReturn(platform, section)
}

// Humanizes returns the human readable plurals of the platform and section tags combined.
//
//   - A plural example: "Windows intros"
//   - A singular example: "a Windows intro"
//
// For singular artifacts and items, use [Humanize].
func (t Tag) Humanizes(section Tag) string {
	switch t {
	case ANSI:
		return pluralANSI(section)
	case Audio:
		return "music, chiptunes, and audio"
	case DataB:
		return pluralDataB(section)
	case DOS:
		return pluralDOS(section)
	case Image:
		return pluralImage(section)
	case Java:
		return Names()[section] + " on java"
	case Linux:
		return Names()[section] + " programs on Linux and Unix"
	case Markup:
		return Names()[section] + " as HTML files"
	case Mac:
		return Names()[section] + " on Macintosh and macOS"
	case PCB:
		return pluralPCB(section)
	case PDF:
		return Names()[section] + " as PDF documents"
	case PHP:
		return Names()[section] + " for any scripting language"
	case Text:
		return pluralText(section)
	case TextAmiga:
		return pluralTextAmiga(section)
	case Video:
		return "videos and animations"
	case Windows:
		return pluralWindows(section)
	}
	return genericPlural(t, section)
}

func check(platform, section Tag) string {
	name := platform.String()
	if !IsPlatform(name) {
		return "unknown platform tag: " + name
	}

	name = section.String()
	if !IsCategory(name) {
		return "unknown section tag: " + name
	}

	return ""
}

func genericReturn(platform, section Tag) string {
	return Determiner()[platform] + " " + Names()[platform] + " " + Names()[section]
}

// sections runs first and always returns a string.
// Meaning it will always humanize any matched sections or categories,
// regardless of the platform.
//
// The t Tag should always be a section.
func (t Tag) sections(platform Tag) string {
	switch t {
	case Bust:
		return platform.humanizeBust()
	case News:
		return platform.humanizeNews()
	case Restrict:
		return platform.humanizeRestrict()
	default:
		return ""
	}
}

// humanizeBust humanizes busts and takedown topics.
func (t Tag) humanizeBust() string {
	const about = "about a bust or takedown"
	switch t {
	case Audio:
		return "audio " + about
	case Video:
		return "video " + about
	case Image:
		return "scan or photo " + about
	case TextAmiga, Text:
		return "bust or takedown text"
	default:
		return "a " + Names()[t] + " takedown notice"
	}
}

// humNews humanizes news, mainstream reports, newpaper type-outs, etc.
func (t Tag) humanizeNews() string {
	const news = "an unauthorized reprint of a newspaper article"
	switch t {
	case Console:
		return "a news article about consoles"
	case Image:
		return "a screenshot of an article from a news outlet"
	case Markup:
		return news + " as HTML"
	case PDF:
		return news + " as PDF"
	case Text:
		return news + " as a textfile"
	case TextAmiga:
		return news + " as an amiga textfile"
	default:
		return Determiner()[t] + " " + Names()[t] + " from a news outlet"
	}
}

// humRetrict humanizes restricted documents, Scene insider files, logfiles,
// and internal documentation that were never intended to be public.
func (t Tag) humanizeRestrict() string {
	switch t {
	case ANSI:
		return "an insider ansi textfile"
	case Text:
		return "an insider textfile"
	case TextAmiga:
		return "an insider amiga/console textfile"
	default:
		return "an insider " + Names()[t] + " file"
	}
}

// platform runs second, after section and will either return a string or an empty value.
// Meaning it only humanizes matched sections or categories of the platform.
//
// The t Tag should always be a platform.
func (t Tag) platforms(section Tag) string {
	switch t {
	case ANSI:
		return section.humanizeANSI(t)
	case Audio:
		return section.humanizeAudio(t)
	case Console:
		return section.humanizeConsole(t)
	case DataB:
		return section.humanizeDataB()
	case DOS:
		return section.humanizeDOS(t)
	case Markup:
		return Determiner()[section] + " " + Names()[section] + " in HTML"
	case Image:
		return section.humanizeImage(t)
	case PDF:
		return "a PDF document about " + Names()[section]
	case Text:
		return section.humanizeText(t)
	case TextAmiga:
		return section.humanizeTextAmiga(t)
	case Video:
		return section.humanizeVideo()
	case Windows:
		return section.humanizeWindows(t)
	default:
		return genericReturn(t, section)
	}
}

// humanizeANSI humanizes textfiles encoded as ANSI.
func (t Tag) humanizeANSI(platform Tag) string {
	switch t {
	case BBS:
		return "an ansi BBS advert"
	case Drama:
		return "a dramatic text in ansi format"
	case Ftp:
		return "an FTP site advert in ansi format"
	case Logo:
		return "an ansi logo"
	case Nfo:
		return "a release text in ansi format"
	case Pack:
		return "a filepack of ansi files"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeAudio humanizes audio files, sound samples, music, voice recordings.
func (t Tag) humanizeAudio(platform Tag) string {
	switch t {
	case Intro:
		return "chiptune or scene music"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeConsole humanizes a file for a video game console.
func (t Tag) humanizeConsole(platform Tag) string {
	switch t {
	case BBS:
		return "a BBS ad on console"
	case Demo:
		return "a demo on console"
	case Drama:
		return "console scene drama"
	case ForSale:
		return "an ad selling console hardware"
	case Guide:
		return "a console guide or help"
	case Logo:
		return "console scene logo"
	case Mag:
		return "console scene magazine"
	case News:
		return "" // unused
	case Nfo:
		return "a console release"
	case Rule:
		return "console scene stardards"
	case Tool:
		return "a console utility"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeDataB humanizes a database file.
func (t Tag) humanizeDataB() string {
	switch t {
	case Nfo:
		return "a database of releases"
	default:
		return Determiner()[t] + " " + Names()[t] + " database"
	}
}

// humanizeDOS humanizes files and applications that are about or intended for
// Microsoft's MS-DOS operating system.
func (t Tag) humanizeDOS(platform Tag) string {
	switch t {
	case AppleII:
		return "a " + msDos + " tool to work with the Apple II"
	case BBS:
		return "a BBStro on " + msDos
	case Demo:
		return "a demo on " + msDos
	case ForSale:
		return "an advertisement on " + msDos
	case Guide:
		return "a how-to guide for a " + msDos + " program"
	case GameHack:
		return "a trainer or hack on " + msDos
	case Install:
		return "a " + msDos + " installer"
	case Intro:
		return "an intro on " + msDos
	case Job:
		return "an job advert or job generator on " + msDos
	case Pack:
		return "a filepack of " + msDos + " programs"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeImage humanizes sceenshots, scanned images, pictures and drawings, and photos.
func (t Tag) humanizeImage(platform Tag) string {
	switch t {
	case AppleII:
		return "an Apple II screen"
	case BBS:
		return "a BBS advert image"
	case ForSale:
		return "an advertisement image"
	case Guide:
		return "a game map or technical drawing"
	case Pack:
		return "a filepack of images"
	case Proof:
		return "a proof of release photo"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeText humanizes textfiles, sometimes called ascii text, plain text, or raw text.
func (t Tag) humanizeText(platform Tag) string {
	switch t {
	case AtariST:
		return "an Atari ST textfile"
	case AppleII:
		return "an Apple II textfile"
	case BBS:
		return "a BBS advert textfile"
	case Drama:
		return "a dramatic textfile"
	case ForSale:
		return "an advertising textfile"
	case Ftp:
		return "a FTP site advert textfile"
	case Logo:
		return "a text logo or brand"
	case Job:
		return "a job advert or job application textfile"
	case Mag:
		return "a magazine textfile"
	case Nfo:
		return "a release textfile"
	case Pack:
		return "a filepack of textfiles"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeTextAmiga humanizes textfiles, sometimes called ascii text, plain text, or raw text.
// However, these use the more limited Latin-1 text encoding, aka ISO/IEC 8859-1.
//
// These texts are displayed using a unique Topaz font that originates
// on the Commodore Amiga microcomputer platform.
func (t Tag) humanizeTextAmiga(platform Tag) string {
	const use = "using amiga text"
	const the = "the amiga or a console"
	switch t { //nolint:exhaustive
	case Announcement:
		return "an announcement about the " + the
	case BBS:
		return "a BBS advert " + use
	case Drama:
		return "a drama textfile concerning " + the
	case ForSale:
		return "an advertising textfile for " + the
	case Ftp:
		return "a FTP site advert " + use
	case Logo:
		return "a logo " + use
	case Job:
		return "a job advert or job application textfile for " + the
	case Mag:
		return "a magazine " + use
	case Nfo:
		return "an amiga or console release textfile"
	default:
		return genericReturn(platform, t)
	}
}

// humanizeVideo humanizes animations and video captures or recordings.
func (t Tag) humanizeVideo() string {
	switch t {
	case ForSale, Logo, Intro:
		return "a bumper video"
	}
	return Determiner()[t] + " " + Names()[t] + " video"
}

// humanizeWindows humanizes files and applications that are about or intended for
// Microsoft's Windows operating system.
//
// There is no acknowledgement of the many different Windows generations or CPU platforms.
func (t Tag) humanizeWindows(platform Tag) string {
	switch t { //nolint:exhaustive
	case Demo:
		return "a demo on Windows"
	case Install:
		return "a Windows installer"
	case Intro:
		return "a Windows intro"
	case Job:
		return "a trial crackme on Windows"
	case Pack:
		return "a filepack of Windows programs"
	default:
		return genericReturn(platform, t)
	}
}

func genericPlural(platform, section Tag) string {
	sections := func(section Tag) string {
		switch section { //nolint:exhaustive
		case BBS:
			return "BBS adverts"
		case Bust:
			return "busted releasers, sites, and sceners"
		case Drama:
			return "drama between releasers or individuals"
		case ForSale:
			return "adverts for releasers or individuals"
		case Ftp:
			return "FTP site adverts"
		case Job:
			return "job adverts or new roles"
		case GameHack:
			return "game trainers or hacks"
		case Guide:
			return "guides, tutorials, and how-to's"
		case Mag:
			return "magazine issues or ads"
		case News:
			return "reprinted articles from media outlets"
		case NfoTool:
			return "nfo file editors or tools"
		case Restrict:
			return "insider or restricted content"
		case Tool:
			return "software tools by the scene"
		}
		return Names()[section] + "s"
	}
	everything := platform < 0 && section < 0
	if everything {
		return "all files"
	}
	if platform < 0 {
		return sections(section)
	}
	if section < 0 {
		return Names()[platform] + "s"
	}
	return fmt.Sprintf("%ss %ss", Names()[platform], Names()[section])
}

func pluralANSI(section Tag) string {
	switch section { //nolint:exhaustive
	case BBS:
		return "BBS ansi adverts"
	case Ftp:
		return "FTP ansi adverts"
	case Logo:
		return "logos in an ansi format"
	case Nfo:
		return "infos in an ansi format"
	case Pack:
		return "filepacks of ansi files"
	default:
		return "texts in an ansi format"
	}
}

func pluralDataB(section Tag) string {
	switch section { //nolint:exhaustive
	case Nfo:
		return "databases of releases"
	default:
		return Names()[section] + " databases"
	}
}

func pluralImage(section Tag) string {
	switch section { //nolint:exhaustive
	case BBS:
		return "BBS advert images"
	case ForSale:
		return "image advertisements"
	case Guide:
		return "how-to guide or help images"
	case Pack:
		return "filepacks of images"
	case Proof:
		return "photos used to prove a release"
	default:
		return "images, pictures, and photos"
	}
}

func pluralPCB(section Tag) string {
	switch section { //nolint:exhaustive
	case BBS:
		return "PCBoard color text files"
	case Tool:
		return "PCBoard script or executable (PPL/PPE)"
	default:
		return "PCBoard bulletin board files"
	}
}

func pluralText(section Tag) string {
	switch section { //nolint:exhaustive
	case AtariST:
		return "textfiles for the Atari ST"
	case AppleII:
		return "textfiles for the Apple II"
	case BBS:
		return "BBS text adverts"
	case Drama:
		return "drama textfiles"
	case ForSale:
		return "textfile adverts"
	case Ftp:
		return "textfile adverts for FTP sites"
	case Mag:
		return "magazine textfiles"
	case Nfo:
		return "release textfiles"
	case Pack:
		return "filepacks of textfiles"
	case Restrict:
		return "restricted or insider textfiles"
	default:
		return Names()[section] + " textfiles"
	}
}

func pluralTextAmiga(section Tag) string {
	const noun = "amiga/console text"
	switch section { //nolint:exhaustive
	case BBS:
		return "BBS " + noun + " adverts"
	case ForSale:
		return noun + "file adverts"
	case Ftp:
		return noun + " adverts for FTP sites"
	case Mag:
		return noun + " magazines"
	case Nfo:
		return noun + " infos"
	case Restrict:
		return "restricted or insider " + noun + "s"
	default:
		return Names()[section] + noun + "s"
	}
}

func pluralWindows(section Tag) string {
	switch section { //nolint:exhaustive
	case Demo:
		return "demos on Windows"
	case Install:
		return "Windows installers"
	case Intro:
		return "Windows intros"
	case Job:
		return "\"CrackMe\" challenges on Windows"
	case Pack:
		return "filepacks of Windows programs"
	default:
		return Names()[section] + " on Windows"
	}
}

func pluralDOS(section Tag) string {
	switch section { //nolint:exhaustive
	case BBS:
		return "BBS intro adverts"
	case Demo:
		return "demos on " + msDos
	case ForSale:
		return "advertisements on " + msDos
	case GameHack:
		return "trainers or hacks on " + msDos
	case Guide:
		return "how-to guide for a " + msDos + " program"
	case Install:
		return msDos + " installers"
	case Intro:
		return "intros on " + msDos
	case Job:
		return "job or generators on " + msDos
	case Pack:
		return "filepacks of " + msDos + " programs"
	default:
		return fmt.Sprintf("%s on %s", Names()[section], msDos)
	}
}

// TagData holds the tag information.
type TagData struct {
	URI   string // URI is a unique URL slug for the tag.
	Name  string // Name is the tags displayed title.
	Info  string // Info is a short description of the tag.
	Count int    // Count is the results of file count query for the tag.
}

// T is a lockable collection of tags, to stop potential race conditions
// when writing to the map containing the tagdata list.
type T struct {
	List []TagData
	Mu   sync.RWMutex
}

// ByName returns the data of the named tag.
// It requires the database to be connected to build the tags if they have not already been.
func (t *T) ByName(name string) (TagData, error) {
	const format = "tags by name %w"
	if t.List == nil {
		return TagData{
			URI:   "",
			Name:  "",
			Info:  "",
			Count: 0,
		}, fmt.Errorf(format, ErrNoTags)
	}
	for val := range slices.Values(t.List) {
		if strings.EqualFold(val.Name, name) {
			return val, nil
		}
	}
	return TagData{
		URI:   "",
		Name:  "",
		Info:  "",
		Count: 0,
	}, nil
}

// Build the tags and collect the statistical data sourced from the database.
func (t *T) Build(ctx context.Context, exec boil.ContextExecutor) (err error) { //nolint:nonamedreturns
	const format = "tags builder %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	if InvalidExec(exec) {
		err = errors.Join(err, fmt.Errorf(format, "", ErrNoDB))
	}
	t.List = make([]TagData, LastPlatform+1)
	i := -1
	for key, val := range URIs() {
		i++
		count64, cErr := counter(ctx, exec, key)
		if err != nil {
			err = errors.Join(err, fmt.Errorf(format, "counter", cErr))
		}
		count := int(count64)
		t.Mu.Lock()
		t.List[i] = TagData{
			URI:   val,
			Name:  Names()[key],
			Info:  Infos()[key],
			Count: count,
		}
		t.Mu.Unlock()
	}
	return err
}

// counter counts the number of files with the tag.
func counter(ctx context.Context, exec boil.ContextExecutor, t Tag) (int64, error) {
	const format = "tags counter %s: %w"
	clause := "section = ?"
	if t >= FirstPlatform {
		clause = "platform = ?"
	}
	sum, err := models.Files(qm.Where(clause, URIs()[t])).Count(ctx, exec)
	if err != nil {
		return -1, fmt.Errorf(format, "could not count the tag", err)
	}
	return sum, nil
}

// IsCategory returns true if the named tag is a category.
func IsCategory(name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return false
	}
	tag, ok := nameToTag[name]
	return ok && tag >= FirstCategory && tag <= LastCategory
}

// IsPlatform returns true if the named tag is a platform.
func IsPlatform(name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return false
	}
	tag, ok := nameToTag[name]
	return ok && tag >= FirstPlatform && tag <= LastPlatform
}

// IsTag returns true if the named tag is a category or platform.
func IsTag(name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return false
	}
	_, ok := nameToTag[name]
	return ok
}

// nameToTag is a reverse lookup map from tag name to Tag for O(1) lookups.
//
//nolint:gochecknoglobals
var nameToTag = func() map[string]Tag {
	const size = 44
	m := make(map[string]Tag, size)
	for _, tag := range List {
		m[strings.ToLower(tag.String())] = tag
	}
	return m
}()

// IsText returns true if the named tag is a raw or plain text category.
func IsText(name string) bool {
	name = strings.TrimSpace(name)
	return strings.EqualFold(name, Text.String()) ||
		strings.EqualFold(name, TextAmiga.String()) ||
		strings.EqualFold(name, Markup.String())
}

// OSTags returns the tags that flag an operating system.
func OSTags() [5]string {
	return [5]string{
		URIs()[DOS],
		URIs()[Java],
		URIs()[Linux],
		URIs()[Windows],
		URIs()[Mac],
	}
}

// InvalidExec returns true if the database context executor is invalid such as nil.
func InvalidExec(exec boil.ContextExecutor) bool {
	v := reflect.ValueOf(exec)
	switch v.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return true
		}
		return false
	}
	return true
}
