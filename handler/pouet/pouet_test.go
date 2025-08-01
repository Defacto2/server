package pouet_test

import (
	"testing"

	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

// Set to true to test against the remote servers.
const testRemoteServers = false

func TestPlatforms(t *testing.T) {
	t.Parallel()
	p := pouet.Platforms{
		DosGus: pouet.Platform{
			Name: "DOS/GUS",
			Slug: "msdosgus",
		},
	}
	be.Equal(t, "msdosgus", p.String())
	be.True(t, p.Valid())
}

func TestType(t *testing.T) {
	t.Parallel()
	var pt pouet.Type = "demo"
	var fd pouet.Type = "fastdemo"
	var prods pouet.Types = []pouet.Type{pt, fd}
	be.True(t, prods.Valid())
	be.Equal(t, "demo, fastdemo", prods.String())
}

func TestResponseGet(t *testing.T) {
	t.Parallel()
	r := pouet.Response{}
	_, err := r.Get(0)
	be.Err(t, err)
	// this pings a remote server, so it is disabled.
	if testRemoteServers {
		_, err = r.Get(1)
		be.Err(t, err, nil)
	}
}

func TestPouet(t *testing.T) {
	t.Parallel()
	p := pouet.Production{}
	_, err := p.Get(0)
	be.Err(t, err)
	// this pings a remote server, so it is disabled.
	if testRemoteServers {
		_, err = p.Get(1)
		be.Err(t, err, nil)
	}
}

func TestVotes(t *testing.T) {
	t.Parallel()
	v := pouet.Votes{}
	err := v.Votes(0)
	be.Err(t, err)
	// this pings a remote server, so it is disabled.
	if testRemoteServers {
		err = v.Votes(1)
		be.Err(t, err, nil)
	}
}

func TestStars(t *testing.T) {
	type args struct {
		up   uint64
		meh  uint64
		down uint64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"1 up", args{1, 0, 0}, 5},
		{"1 meh", args{0, 1, 0}, 3},
		{"1 down", args{0, 0, 1}, 1},
		{"2 below avg", args{0, 1, 1}, 2},
		{"1s", args{1, 1, 1}, 3},
		{"1,1,0", args{1, 1, 0}, 4},
		{"2,1,0", args{2, 1, 0}, 4.5},
		{"3,1,0", args{3, 1, 0}, 4.5},
		{"7,1,0", args{7, 1, 0}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := pouet.Stars(tt.args.up, tt.args.meh, tt.args.down)
			be.True(t, f == tt.want)
		})
	}
}

func TestValids(t *testing.T) {
	t.Parallel()
	be.True(t, !pouet.PlatformsValid(""))
	be.True(t, pouet.PlatformsValid("msdos"))
	be.True(t, pouet.TypesValid(""))
	be.True(t, !pouet.TypesValid("fastdemo"))
	be.True(t, pouet.TypesValid("msdos,random,placeholder"))
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	pp := pouet.Production{}
	a, b := pp.Releasers()
	be.True(t, a == "")
	be.True(t, b == "")
	x, y, z := pp.Released()
	be.True(t, x == 0)
	be.True(t, y == 0)
	be.True(t, z == 0)
}

func TestPlatformType(t *testing.T) {
	t.Parallel()
	pp := pouet.Production{}
	a, b := pp.PlatformType()
	be.Equal(t, tags.Tag(-1), a)
	be.Equal(t, tags.Intro, b)
}
