//nolint:wrapcheck
package app

// Package file context_static.go contains handlers for static and editorial pages.

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/nils"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
)

const (
	apps        = "apps"
	areacodes   = "areacodes"
	brokentexts = "brokentexts"
	compression = "compression"
	credits     = "credits"
	fixes       = "fixes"
	history     = "history"
	index       = "index"
	interview   = "interview"
	newx        = "new"
	policy      = "policy"
	terms       = "terms"
	thanks      = "thanks"
	thescene    = "thescene"
	titles      = "titles"
)

// APIInfo is the handler for the APIInfo end-user helper page.
func APIInfo(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    "api-info",
		errContext:  "context api info",
		title:       "API Information",
		h1:          "RESTful API",
		logo:        "application programming interface",
		lead:        "Basic information on how to use the Defacto2 API.",
		description: "Read-only RESTful JSON API documentation, endpoints, and examples for Defacto2.",
	})
}

// Apps is the handler for the modern applications and tools page.
func Apps(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:   apps,
		errContext: "apps handler context",
		title:      "Use current apps",
		h1:         "Modern Applications and Tools",
		logo:       "Software suggestions",
		lead: "Here are some software suggestions and Windows, Linux, macOS tools for running out-of-date programs " +
			"and using legacy media formats.",
		description: "Software and application suggestions for using the historic artifacts and " +
			"file downloads on modern systems.",
	})
}

// Areacodes is the handler for the BBS and telephone area codes page.
func Areacodes(sl *slog.Logger, c *echo.Context) error {
	const format = "areacodes context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "BBS and telephone area codes"
	const descr = "Lookup and list the North American Numbering Plan (NANP) area codes in common use until 1994."
	const leadr = "North American Numbering Plan (+1-XXX) telephone area codes until 1994."
	const logo = "BBS area codes"

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = logo
	data["h1"] = logo
	data["lead"] = leadr
	data["telephonecodes"] = areacode.AreaCodes()
	data["territories"] = areacode.Regions()
	data["abbreviations"] = areacode.Abbreviations()
	data["areaCodeEntries"] = BuildAreaCodeEntries()
	data["regionEntries"] = BuildAreaRegionEntries()
	data["abbrEntries"] = BuildAreaAbbrEntries()

	if err := c.Render(http.StatusOK, areacodes, data); err != nil {
		return InternalErr(sl, c, areacodes, err)
	}
	return nil
}

// BrokenTexts is the handler for the Broken texts page.
func BrokenTexts(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:   brokentexts,
		errContext: "broken texts context",
		title:      "Broken texts!?",
		h1:         "Broken text files",
		logo:       "Broken text files and NFOs",
		lead: "Unfortunately, there are large numbers of incomplete, inaccurate, " +
			"or corrupted information texts (NFOs). While we'd prefer to offer a pristine copy of a Scene text, " +
			"hosting a broken text is more useful than offering nothing.",
		description: "Learn why there are broken encodings, unreadable texts and corrupted nfo files.",
	})
}

// Compression is the handler for information on historic compression tools page.
func Compression(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:   compression,
		errContext: "compression context",
		title:      "Compression and archiving formats",
		h1:         "File compression formats",
		logo:       "Old archives",
		lead: "Compression and archiving formats of the 1980s were evolving by the month, " +
			"and today, are hard to parse.",
		description: "Old file archives and compression methods used in the 1980s on the PC.",
	})
}

// Credits is the handler for the credits and acknowledgments page.
func Credits(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    credits,
		errContext:  "credits context",
		title:       "Credits",
		h1:          "Credits",
		logo:        "Credits",
		lead:        "The people, open-source projects, and scene contributors that power Defacto2.",
		description: "Credits, core technologies, and scene contributors behind Defacto2.",
	})
}

// Fixes is the handler for the problems and fixes page.
func Fixes(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    fixes,
		errContext:  "fixes for programs context",
		title:       "Problems and fixes",
		h1:          "Common problems and fixes",
		logo:        "Problems and fixes",
		lead:        "Shrinker dispatcher or runtime 200 errors, or missing NPMOD32.DLL or D3DRM.DLL files?",
		description: "Fix common errors found in applications and tools authored by the Scene for Windows and DOS.",
	})
}

// History is the handler for the History page.
func History(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:   history,
		errContext: "history context",
		title:      "The history of the brand",
		h1:         "The history of the brand",
		logo:       "The history of Defacto",
		carousel:   `#carouselDf2Artpacks`,
		lead: "In the past, alternative iterations of the name have included" +
			" De Facto, DF, DeFacto, Defacto II, Defacto 2, and the defacto2.com domain.",
		description: "Learn about the many iterations of Defacto2 and the original DeFacto magazine from 1996.",
	})
}

// Index is the handler for the Home page.
func Index(sl *slog.Logger, c *echo.Context) error {
	const format = "index context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Introduction and milestones"
	const descr = "Preserving the historic and defunct software cracking, wares, and Warez Scene subcultures on obsolete personal computers."
	const h1 = "The subcultures of obsolete microcomputers"

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data[canonical] = "/"
	data["h1"] = ""
	data["heroTitle"] = h1
	if meta := archiveMeta(); meta != "" {
		data["archiveMeta"] = meta
	}
	data["indexJS"] = true
	data["milestones"] = Collection()
	{
		// get the given name of the signed in session
		s, err := session.Get(sess.Name, c)
		if err == nil {
			if givenName, givenExists := s.Values["givenName"]; givenExists {
				if name, nameStr := givenName.(string); nameStr && name != "" {
					data["heroTitle"] = "Welcome, " + name
				}
			}
		}
	}

	if err := c.Render(http.StatusOK, index, data); err != nil {
		return InternalErr(sl, c, index, err)
	}
	return nil
}

func archiveMeta() string {
	parts := make([]string, 0, 3)
	if Caching.RecordCount > 0 {
		parts = append(parts, humanize.Comma(Caching.RecordCount)+" artifacts")
	}
	if Caching.TotalBytes > 0 {
		parts = append(parts, helper.ByteCountFloat(Caching.TotalBytes))
	}
	if Caching.MinYear > 0 && Caching.MaxYear > 0 {
		parts = append(parts, fmt.Sprintf("%d–%d", Caching.MinYear, Caching.MaxYear))
	}
	return strings.Join(parts, " · ")
}

// Interview is the handler for the People Interviews page.
func Interview(sl *slog.Logger, c *echo.Context) error {
	const format = "interview context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Interviews with former Sceners"
	const descr = "A collection of historical interviews and discussions with former members of the Scene."
	const leadr = "Here is a centralized page for the discussions and unedited" +
		" interviews with former sceners, crackers, and demo makers."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = title
	data["h1"] = title
	data["lead"] = leadr
	items := Interviewees()
	data["interviews"] = items
	data["interviewCount"] = InterviewCount(items)

	if err := c.Render(http.StatusOK, interview, data); err != nil {
		return InternalErr(sl, c, interview, err)
	}
	return nil
}

// New is the handler for the what is new page.
func New(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    newx,
		errContext:  "new context",
		title:       "New stuff",
		h1:          "What is new?",
		logo:        "New stuff",
		lead:        "This quaint page does not appeal to algorithms, so who will see it?",
		description: "What is new on the Defacto2 website?",
		carousel:    "#carouselWhatsNew",
		noindex:     true, // apply noindex to what's new, so we don't have to worry about <a href rel="nofollow">
	})
}

// Policy is the handler for the privacy and copyright policy page.
func Policy(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    policy,
		errContext:  "policy context",
		title:       "Privacy & copyright",
		h1:          "Privacy & copyright",
		logo:        "Privacy & copyright",
		lead:        "How we handle visitor privacy and how Defacto2 content may be reused.",
		description: "Defacto2 privacy policy and Creative Commons copyright licensing.",
	})
}

// Terms is the handler for the glossary of terms page.
func Terms(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    terms,
		errContext:  "glossary of terms context",
		title:       "Common terms",
		h1:          "Glossary of common terms",
		logo:        "Glossary of terms",
		lead:        "A glossary of unique and common terms used in The Scene.",
		description: "This is a list of the unique and common terms used in the scene.",
	})
}

// Thanks is the handler for the Thanks page.
func Thanks(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:   thanks,
		errContext: "thanks context",
		title:      "Thank you!",
		h1:         "Thank you!",
		lead: "Thanks to the hundreds of people who have contributed to" +
			" Defacto2 over the decades with file submissions, " +
			"hard drive donations, interviews, corrections, artwork, and monetary contributions!",
		description: "A special thanks to the hundreds of contributors and the thousands of contributions.",
	})
}

// TheScene is the handler for the The Scene page.
func TheScene(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:   thescene,
		errContext: "the scene context",
		title:      "The Scene",
		h1:         "The Scene",
		logo:       "The underground",
		lead: "The Scene is a broad church of people and online communities that is collectively grouped;" +
			" it is a subculture of niche activities using personal computers, " +
			"where the participants share creations and exchange ideas.",
		description: "A short introduction on The Scene, the online subcultures and its underground origins.",
	})
}

// Titles is the handler for the Titles page.
func Titles(sl *slog.Logger, c *echo.Context) error {
	return renderStatic(sl, c, staticPage{
		template:    titles,
		errContext:  "titles context",
		title:       "Titles",
		h1:          "Titles and naming are important",
		logo:        "Titles",
		lead:        "A concise guide to naming software, games, and Scene releases cleanly.",
		description: "Naming and title guidelines for Defacto2 uploads and archive entries.",
	})
}

// Website is the handler for the websites page.
// Open is the ID of the accordion section to open.
func Website(sl *slog.Logger, c *echo.Context, open string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("website context: %w", err)
	}

	const title = "Websites"
	const logo = "Videos, Books, Films, Sites, Podcasts"
	data := empty(c)
	data["title"] = title
	data["h1"] = title
	data["lead"] = "A curated collection of videos, books, films, websites, podcasts, and AMAs about The Scene."
	data["logo"] = logo
	data["description"] = "Our curated collection of " + strings.ToLower(logo) + ", with topics about the Scene."

	accordion := List()
	// Open the accordion section.
	closeAll := true
	for i, s := range accordion {
		if s.ID == open || open == "" {
			s.Open = true
			closeAll = false
			accordion[i] = s
			if open == "" {
				continue
			}
			break
		}
	}
	if closeAll {
		data["noindex"] = true
		data["title"] = title + " categories"
	}

	// If a section was requested but not found, return a 404.
	if open != "hide" && closeAll {
		return StatusErr(sl, c, http.StatusNotFound, open)
	}

	// Render the page.
	data["accordion"] = accordion
	if err := c.Render(http.StatusOK, websites, data); err != nil {
		return InternalErr(sl, c, "render open website, "+open, err)
	}
	return nil
}
