package demozoo_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestGet(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	prod := demozoo.Production{}
	_, got := prod.Get(ctx, "", -1)
	be.Err(t, got, demozoo.ErrID)

	if testRemoteServers {
		auth := "" // NOTE: required for remote access
		_, got = prod.Get(ctx, auth, 1)
		be.Err(t, got, demozoo.ErrSuccess)
	}
}

func TestFind(t *testing.T) {
	t.Parallel()

	be.Equal(t, demozoo.Find("defacto2"), demozoo.GroupID(10000))
	be.Equal(t, demozoo.Find("destined-masters-of-zines"), demozoo.GroupID(123572))
	be.Equal(t, demozoo.Find("notfound"), demozoo.GroupID(0))
}

func TestCopy(t *testing.T) {
	t.Parallel()

	all := demozoo.Copy()
	be.True(t, len(all) > 500)
	be.Equal(t, all["defacto2"], demozoo.GroupID(10000))
	for uri, id := range all {
		if !name.Valid(name.Path(uri)) {
			t.Errorf("%q: invalid URI slug", uri)
		}
		if id <= 0 {
			t.Errorf("%q: GroupID %d must be positive", uri, id)
		}
	}
	// Mutating the returned map must not affect the backing map.
	all["defacto2"] = 0
	be.Equal(t, demozoo.Find("defacto2"), demozoo.GroupID(10000))
}

func TestExternalLinks(t *testing.T) {
	t.Parallel()

	type extLink = struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}

	d := demozoo.Production{}
	be.Equal(t, d.PouetProd(), 0)
	be.Equal(t, d.GithubRepo(), "")
	be.Equal(t, d.YouTubeVideo(), "")

	d.ExternalLinks = []extLink{
		{LinkClass: "class1", URL: "http://example.com/1"},
		{LinkClass: "PouetProduction", URL: "://bad"},
		{LinkClass: "PouetProduction", URL: "http://example.com/1"},
		{LinkClass: "GithubRepo", URL: "://bad"},
		{LinkClass: "GithubRepo", URL: "https://gitlab.com/Defacto2/server"},
		{LinkClass: "YoutubeVideo", URL: "://bad"},
		{LinkClass: "YoutubeVideo", URL: "https://vimeo.com/watch?v=abc"},
		{LinkClass: "YoutubeVideo", URL: "https://youtube.com/embed/abc"},
	}
	be.Equal(t, d.PouetProd(), 0)
	be.Equal(t, d.GithubRepo(), "")
	be.Equal(t, d.YouTubeVideo(), "")

	d.ExternalLinks = append(d.ExternalLinks,
		extLink{LinkClass: "PouetProduction", URL: "https://www.pouet.net/prod.php?which=71562"},
		extLink{LinkClass: "GithubRepo", URL: "https://github.com/Defacto2/server"},
		extLink{LinkClass: "YoutubeVideo", URL: "https://www.youtube.com/watch?v=x6QrKsBOERA"},
	)
	be.Equal(t, d.PouetProd(), 71562)
	be.Equal(t, d.GithubRepo(), "/Defacto2/server")
	be.Equal(t, d.YouTubeVideo(), "x6QrKsBOERA")
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("nil reader", func(t *testing.T) {
		t.Parallel()

		var prod demozoo.Production
		be.Err(t, prod.Unmarshal(nil), nil)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		var prod demozoo.Production
		be.Err(t, prod.Unmarshal(strings.NewReader(`{bad`)))
	})

	t.Run("invalid id", func(t *testing.T) {
		t.Parallel()

		var prod demozoo.Production
		be.Err(t, prod.Unmarshal(strings.NewReader(`{"id":0,"title":"Zero"}`)), demozoo.ErrID)
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		var prod demozoo.Production
		err := prod.Unmarshal(strings.NewReader(`{"id":123,"title":"Second Reality","release_date":"1993-08-01"}`))
		be.Err(t, err, nil)
		be.Equal(t, prod.ID, 123)
		be.Equal(t, prod.Title, "Second Reality")
		be.Equal(t, prod.ReleaseDate, "1993-08-01")
	})
}

func TestSuperType(t *testing.T) {
	t.Parallel()

	type item = struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}
	makeProd := func(platID, typeID int) demozoo.Production {
		var p demozoo.Production
		if platID > 0 {
			p.Platforms = []item{{ID: platID}}
		}
		if typeID > 0 {
			p.Types = []item{{ID: typeID}}
		}
		return p
	}

	tests := []struct {
		name     string
		prod     demozoo.Production
		wantPlat tags.Tag
		wantSect tags.Tag
	}{
		{"empty", demozoo.Production{}, -1, -1},
		{"windows demo", makeProd(1, demozoo.Demo), tags.Windows, tags.Demo},
		{"dos intro 64k", makeProd(4, demozoo.Intro64K), tags.DOS, tags.Intro},
		{"linux diskmag", makeProd(7, demozoo.DiskMag), tags.Linux, tags.Mag},
		{"mac tool", makeProd(10, demozoo.Tool), tags.Mac, tags.Tool},
		{"macintosh cracktro", makeProd(94, demozoo.Cracktro), tags.Mac, tags.Intro},
		{"console game", makeProd(22, demozoo.Game), tags.Console, tags.Demo},
		{"snes pack", makeProd(34, demozoo.ProductionPack), tags.Console, tags.Pack},
		{"browser bbstro", makeProd(12, demozoo.BBStro), tags.Markup, tags.BBS},
		{"textmag", makeProd(0, demozoo.TextMag), tags.Text, tags.Mag},
		{"musicdisk", makeProd(0, demozoo.MusicDisk), tags.Audio, tags.Pack},
		{"music", makeProd(0, demozoo.Music), tags.Audio, tags.Intro},
		{"bbsdoor", makeProd(0, demozoo.BBSDoor), tags.PCB, tags.Tool},
		{"graphics", makeProd(0, 23), tags.Image, tags.Logo},
		{"ascii", makeProd(0, 24), tags.Text, tags.Logo},
		{"pack ascii", makeProd(0, 25), tags.Text, tags.Pack},
		{"ansi", makeProd(0, 26), tags.Text, tags.Logo},
		{"exegfx dos", makeProd(4, 27), tags.DOS, tags.Logo},
		{"artpack", makeProd(0, 51), tags.Image, tags.Pack},
		{"chipmusic", makeProd(0, 29), tags.Audio, tags.Intro},
		{"exemusic dos", makeProd(4, 31), tags.DOS, tags.Intro},
		{"musicpack", makeProd(0, 52), tags.Audio, tags.Pack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plat, sect := tt.prod.SuperType()
			be.Equal(t, plat, tt.wantPlat)
			be.Equal(t, sect, tt.wantSect)
		})
	}
}

func TestReleased(t *testing.T) {
	t.Parallel()

	prod := demozoo.Production{}
	y, m, d := prod.Released()
	be.Equal(t, y, int16(0))
	be.Equal(t, m, int16(0))
	be.Equal(t, d, int16(0))

	prod.ReleaseDate = "1993-08-01"
	y, m, d = prod.Released()
	be.Equal(t, y, int16(1993))
	be.Equal(t, m, int16(8))
	be.Equal(t, d, int16(1))
}

func TestGroups(t *testing.T) {
	t.Parallel()

	type author = struct {
		Name     string `json:"name"`
		Releaser struct {
			Name    string `json:"name"`
			IsGroup bool   `json:"is_group"`
		} `json:"releaser"`
	}
	mkAuthor := func(name string, isGroup bool) author {
		var a author
		a.Name = name
		a.Releaser.Name = name
		a.Releaser.IsGroup = isGroup
		return a
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		prod := demozoo.Production{}
		a, b := prod.Groups()
		be.Equal(t, a, "")
		be.Equal(t, b, "")
	})

	t.Run("groups only skips individuals", func(t *testing.T) {
		t.Parallel()

		prod := demozoo.Production{
			Title: "Second Reality",
			Authors: []author{
				mkAuthor("Purple Motion", false),
				mkAuthor("Future Crew", true),
				mkAuthor("Triton", true),
				mkAuthor("Ignored", true),
			},
		}
		a, b := prod.Groups()
		be.Equal(t, a, "FUTURE CREW")
		be.Equal(t, b, "TRITON")
		be.Equal(t, prod.Title, "Second Reality")
	})

	t.Run("bbs title becomes first group", func(t *testing.T) {
		t.Parallel()

		prod := demozoo.Production{
			Title:   "The Cool BBS intro",
			Authors: []author{mkAuthor("Razor 1911", true)},
		}
		a, b := prod.Groups()
		be.Equal(t, a, "COOL BBS")
		be.Equal(t, b, "RAZOR 1911")
		be.Equal(t, prod.Title, "")
	})
}

func TestSite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"empty", "", ""},
		{"no match", "Second Reality", ""},
		{"the bbs", "the cool bbs", "cool BBS"},
		{"bbs", "Cool BBS", "Cool BBS"},
		{"the ftp", "the fast ftp site", "fast FTP"},
		{"ftp", "Elite FTP", "Elite FTP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, demozoo.Site(tt.title), tt.want)
		})
	}
}

func TestReleasers(t *testing.T) {
	t.Parallel()

	type credit = struct {
		Category string `json:"category"`
		Role     string `json:"role"`
		Nick     struct {
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation"`
			Releaser     struct {
				URL     string `json:"url"`
				Name    string `json:"name"`
				ID      int    `json:"id"`
				IsGroup bool   `json:"is_group"`
			} `json:"releaser"`
		} `json:"nick"`
	}
	mkCredit := func(cat, name string, isGroup bool) credit {
		var c credit
		c.Category = cat
		c.Nick.Name = name
		c.Nick.Releaser.IsGroup = isGroup
		return c
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		prod := demozoo.Production{}
		a, b, c, d := prod.Releasers()
		be.Equal(t, len(a), 0)
		be.Equal(t, len(b), 0)
		be.Equal(t, len(c), 0)
		be.Equal(t, len(d), 0)
	})

	t.Run("categorizes individuals and skips groups", func(t *testing.T) {
		t.Parallel()

		prod := demozoo.Production{
			Credits: []credit{
				mkCredit("Text", "Writer", false),
				mkCredit("Code", "Coder", false),
				mkCredit("Graphics", "Artist", false),
				mkCredit("Music", "Musician", false),
				mkCredit("Magazine", "Editor", false),
				mkCredit("Other", "Unknown", false),
				mkCredit("Code", "GroupName", true),
			},
		}
		tx, co, gx, mu := prod.Releasers()
		be.Equal(t, tx, []string{"Writer"})
		be.Equal(t, co, []string{"Coder"})
		be.Equal(t, gx, []string{"Artist"})
		be.Equal(t, mu, []string{"Musician"})
	})
}

func TestCategory(t *testing.T) {
	t.Parallel()

	be.Equal(t, demozoo.TextC.String(), "text")
	be.Equal(t, demozoo.CodeC.String(), "code")
	be.Equal(t, demozoo.GraphicsC.String(), "graphics")
	be.Equal(t, demozoo.MusicC.String(), "music")
	be.Equal(t, demozoo.MagazineC.String(), "magazine")
	be.Equal(t, demozoo.Category(-1).String(), "")
	be.Equal(t, demozoo.Category(99).String(), "")
}
