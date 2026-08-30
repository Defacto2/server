// Package sitemap generates site mappings and a mapping indexes that can be rendered in XML.
// Sitemaps are used to direct search engines how to navigate their site crawls.
// As Defacto2 has 100,000s of hyperlinks with many generated from the database.
// We need to limit the search engines to avoid them wasting their crawl quote on
// duplicated linkage.
//
// Useful links,
//
//   - [Search Central, Learn about sitemaps]
//   - [XML Validator]
//   - [Sitemaps XML protocol]
//
// [Search Central, Learn about sitemaps]: https://developers.google.com/search/docs/crawling-indexing/sitemaps/overview
// [XML Validator]: https://codebeautify.org/xmlvalidator
// [Sitemaps XML protocol]: https://www.sitemaps.org/protocol.html
package sitemap

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
)

const (
	Namespace = "http://www.sitemaps.org/schemas/sitemap/0.9"
	RootURL   = "https://defacto2.net"
	Website   = "sitemap.xml"
	Releaser  = "sitemap-releaser.xml"
	Magazine  = "sitemap-magazine.xml"
	BBS       = "sitemap-bbs.xml"
	FTP       = "sitemap-ftp.xml"

	limit  = 198 // per-page record limit
	urlset = "urlset"
)

// Index is a xml sitemap index file that is a centralized collection
// of all the sitemaps used on the site.
//
// An example output:
//
// <?xml version="1.0" encoding="UTF-8"?>
// <sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
//
//	<sitemap>
//		<loc>https://www.example.com/sitemap1.xml.gz</loc>
//		<lastmod>2024-08-15</lastmod>
//	</sitemap>
//	<sitemap>
//		<loc>https://www.example.com/sitemap2.xml.gz</loc>
//		<lastmod>2022-06-05</lastmod>
//	</sitemap>
//
// </sitemapindex>
//
// See package documentation for links.
type Index struct {
	XMLName xml.Name `xml:"sitemapindex"`
	XMLNS   string   `xml:"xmlns,attr"`
	Maps    []Map
}

// Map is the URL location of the sitemap.
type Map struct {
	XMLName xml.Name `xml:"sitemap"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

// MapIndex generates the sitemap index xml page.
// It must be handled by either the XML or XMLPretty echo contexts.
func MapIndex() *Index {
	locs := [...]string{
		Website,
		Releaser,
		Magazine,
		BBS,
		FTP,
	}

	maps := make([]Map, len(locs))
	for i, loc := range locs {
		maps[i].Loc = RootURL + "/" + loc
	}

	index := &Index{
		XMLName: xml.Name{Space: "", Local: "sitemapindex"},
		XMLNS:   Namespace,
		Maps:    maps,
	}

	return index
}

// Sitemap is an xml sitemap file that links to the all the
// website pages that should be crawled and index by search engines.
//
// An example output:
//
// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
//
//	<url>
//	  <loc>https://www.example.com/foo.html</loc>
//	  <lastmod>2022-06-04</lastmod>
//	</url>
//
// </urlset>
//
// See package documentation for links.
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	Locs    []Loc
}

// Loc is the URL location of the website page.
type Loc struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

// MapSite generates the main sitemap for the website.
// It must be handled by either the XML or XMLPretty echo contexts.
func MapSite(ctx context.Context, sl *slog.Logger, db *sql.DB) *Sitemap {
	const format = "site mapsite: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, err))
	}

	var m model.Summary
	err := m.ByPublic(ctx, db)
	if err != nil {
		sl.Error("site mapsite could not obtain summary by public", slog.Any("error", err))
	}

	sum := m.SumCount.Int64
	pages := (sum + limit - 1) / limit

	staticLocs := [...]string{
		"bbs/year", "ftp", "magazine", "releaser/year", "scener", "website",
		"areacodes", "history", "brokentexts", "apps", "fixes",
		"compression", "terms", "thescene", "thanks", "files/oldest",
	}
	total := len(staticLocs)
	if pages > 1 {
		total += int(pages - 1)
	}

	maps := make([]Loc, 0, total)
	for _, loc := range staticLocs {
		maps = append(maps, Loc{Loc: RootURL + "/" + loc})
	}
	for p := int64(2); p <= pages; p++ {
		maps = append(maps, Loc{Loc: RootURL + "/files/oldest/" + strconv.FormatInt(p, 10)})
	}

	sm := &Sitemap{
		XMLName: xml.Name{Local: urlset},
		XMLNS:   Namespace,
		Locs:    maps,
	}
	return sm
}

// MapReleaser generates the sitemap that links to every releaser page that is public.
// It must be handled by either the XML or XMLPretty echo contexts.
func MapReleaser(ctx context.Context, sl *slog.Logger, db *sql.DB) *Sitemap {
	const format = "site mapreleaser: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, err))
	}

	var r model.Releasers
	if err := model.Oldest.Limit(ctx, db, &r, 0, 0); err != nil {
		sl.Error("site mapreleaser could not obtain releasers using limit", slog.Any("error", err))
	}

	maps := make([]Loc, len(r))
	for i, rel := range r {
		maps[i] = Loc{Loc: RootURL + "/g/" + rel.Unique.URI}
	}

	sm := &Sitemap{
		XMLName: xml.Name{Local: urlset},
		XMLNS:   Namespace,
		Locs:    maps,
	}
	return sm
}

// MapMagazine generates the sitemap that links to every magazine page that is public.
// It must be handled by either the XML or XMLPretty echo contexts.
func MapMagazine(ctx context.Context, sl *slog.Logger, db *sql.DB) *Sitemap {
	const format = "site mapmagazine: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, err))
	}

	var r model.Releasers
	if err := r.Magazine(ctx, db); err != nil {
		sl.Error("site mapmagazine could not obtain magazine publications", slog.Any("error", err))
	}

	maps := make([]Loc, len(r))
	for i, rel := range r {
		maps[i].Loc = RootURL + "/g/" + rel.Unique.URI
	}

	sm := &Sitemap{
		XMLName: xml.Name{Local: urlset},
		XMLNS:   Namespace,
		Locs:    maps,
	}
	return sm
}

// MapBBS generates the sitemap that links to every bbs page that is public.
// It must be handled by either the XML or XMLPretty echo contexts.
func MapBBS(ctx context.Context, sl *slog.Logger, db *sql.DB) *Sitemap {
	const format = "site mapbbs: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, err))
	}

	var r model.Releasers
	if err := model.Prolific.BBS(ctx, db, &r); err != nil {
		sl.Error("site mapbbs could not obtain bulletin boards", slog.Any("error", err))
	}

	maps := make([]Loc, len(r))
	for i, rel := range r {
		maps[i].Loc = RootURL + "/g/" + rel.Unique.URI
	}

	sm := &Sitemap{
		XMLName: xml.Name{Local: urlset},
		XMLNS:   Namespace,
		Locs:    maps,
	}
	return sm
}

// MapFTP generates the sitemap that links to every ftp page that is public.
// It must be handled by either the XML or XMLPretty echo contexts.
func MapFTP(ctx context.Context, sl *slog.Logger, db *sql.DB) *Sitemap {
	const format = "site mapftp: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, err))
	}

	var r model.Releasers
	if err := r.FTP(ctx, db); err != nil {
		sl.Error("site mapftp could not obtain file sites", slog.Any("error", err))
	}

	maps := make([]Loc, len(r))
	for i, rel := range r {
		maps[i].Loc = RootURL + "/g/" + rel.Unique.URI
	}

	sm := &Sitemap{
		XMLName: xml.Name{Local: urlset},
		XMLNS:   Namespace,
		Locs:    maps,
	}
	return sm
}
