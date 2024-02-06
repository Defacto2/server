package zoo_test

import (
	"testing"

	"github.com/Defacto2/server/internal/zoo"
	"github.com/stretchr/testify/assert"
)

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestDemozoo_Get(t *testing.T) {
	t.Parallel()
	prod := zoo.Demozoo{}
	err := prod.Get(-1)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &zoo.ErrID)

	if !testRemoteServers {
		return
	}

	err = prod.Get(1)
	assert.NoError(t, err)
	assert.ErrorAs(t, err, &zoo.ErrSuccess)
}

func TestFind(t *testing.T) {
	t.Parallel()
	prod := zoo.Find("defacto2")
	want := zoo.GroupID(10000)
	assert.Equal(t, want, prod)

	prod = zoo.Find("notfound")
	assert.Equal(t, prod, zoo.GroupID(0))
}

func TestExternalLinks(t *testing.T) {
	t.Parallel()
	d := zoo.Demozoo{}
	assert.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "class1",
		URL:       "http://example.com/1",
	})
	assert.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "http://example.com/1",
	})
	assert.Equal(t, 0, d.PouetProd())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "PouetProduction",
		URL:       "https://www.pouet.net/prod.php?which=71562",
	})
	assert.Equal(t, 71562, d.PouetProd())
	assert.Empty(t, d.GithubRepo())

	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "GithubRepo",
		URL:       "https://github.com/Defacto2/server",
	})
	assert.Equal(t, "/Defacto2/server", d.GithubRepo())

	assert.Equal(t, "", d.YouTubeVideo())
	d.ExternalLinks = append(d.ExternalLinks, struct {
		LinkClass string `json:"link_class"`
		URL       string `json:"url"`
	}{
		LinkClass: "YoutubeVideo",
		URL:       "https://www.youtube.com/watch?v=x6QrKsBOERA",
	})
	assert.Equal(t, "x6QrKsBOERA", d.YouTubeVideo())
}
