// Package sitemap TODO:
package sitemap

// TODO: move to a sitemap pkg and create a range slice for the uris to add.
// they can lookup the last mod of the view template and use that for lastmod.
import (
	"context"
	"database/sql"
	"encoding/xml"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/Defacto2/server/model"
)

const (
	limit     = 198 // per-page record limit
	Namespace = "http://www.sitemaps.org/schemas/sitemap/0.9"
	RootURL   = "https://defacto2.net"
	Website   = "mapwebsite.xml"
	File1     = "mapfile1.txt"    // keep unused for now
	File2     = "mapfile2.txt"    // "
	Releaser  = "mapreleaser.txt" // make xml and fetch from db
	Magazine  = "mapmagazine.txt" // "
	BBS       = "mapbbs.txt"      // "
	FTP       = "mapftp.txt"      // "
)

type Index struct {
	XMLName xml.Name `xml:"sitemapindex"`
	XMLNS   string   `xml:"xmlns,attr"`
	Maps    []Map
}

type Map struct {
	XMLName xml.Name `xml:"sitemap"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

func MapIndex() *Index {
	locs := []string{Website, File1, File2, Releaser, Magazine, BBS, FTP}
	maps := make([]Map, len(locs))
	for i, loc := range slices.All(locs) {
		maps[i].Loc = RootURL + "/" + loc
		maps[i].LastMod = time.Now().Format(time.DateOnly)
	}
	index := &Index{
		XMLNS: Namespace,
		Maps:  maps,
	}
	return index
}

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	Locs    []Loc
}

type Loc struct {
	XMLName xml.Name `xml:"urlset"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

func MapSite(db *sql.DB) *Sitemap {
	ctx := context.Background()
	m := model.Summary{}
	err := m.ByPublic(ctx, db)
	if err != nil {
		panic(err)
		return nil // TODO: handle
	}
	sum := m.SumCount.Int64
	lastPage := math.Ceil(float64(sum) / float64(limit))
	locs := []string{
		//"files/oldest/" + fmt.Sprint(lastPage),
		//"files/oldest/2", // TODO: ??
		//"files/oldest/273",
		"bbs/year",
		"ftp",
		"magazine",
		"releaser/year",
		"scener",
		"website",
		"areacodes",
		"history",
		"brokentexts",
		"thescene",
		"thanks",
		"files/oldest", // page 1
	}
	count := 1
	for count < int(lastPage) {
		count++
		locs = append(locs, "files/oldest/"+strconv.Itoa(count))
	}
	maps := make([]Loc, len(locs))
	for i, loc := range slices.All(locs) {
		maps[i].Loc = RootURL + "/" + loc
		// maps[i].LastMod = time.Now().Format(time.DateOnly)
	}
	sm := &Sitemap{
		XMLNS: Namespace,
		Locs:  maps,
	}
	return sm
}
