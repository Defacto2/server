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

	tests := []struct {
		name      string
		platforms pouet.Platforms
		wantStr   string
		wantValid bool
	}{
		{
			name:      "empty",
			platforms: pouet.Platforms{},
			wantStr:   "",
			wantValid: false,
		},
		{
			name: "dos gus",
			platforms: pouet.Platforms{
				DosGus: pouet.Platform{Name: "DOS/GUS", Slug: "msdosgus"},
			},
			wantStr:   "msdosgus",
			wantValid: true,
		},
		{
			name: "msdos",
			platforms: pouet.Platforms{
				MSDos: pouet.Platform{Name: "MS-Dos", Slug: "msdos"},
			},
			wantStr:   "msdos",
			wantValid: true,
		},
		{
			name: "windows",
			platforms: pouet.Platforms{
				Windows: pouet.Platform{Name: "Windows", Slug: "windows"},
			},
			wantStr:   "windows",
			wantValid: true,
		},
		{
			name: "all",
			platforms: pouet.Platforms{
				DosGus:  pouet.Platform{Name: "DOS/GUS", Slug: "msdosgus"},
				MSDos:   pouet.Platform{Name: "MS-Dos", Slug: "msdos"},
				Windows: pouet.Platform{Name: "Windows", Slug: "windows"},
			},
			wantStr:   "msdosgus, msdos, windows",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, tt.platforms.String(), tt.wantStr)
			be.Equal(t, tt.platforms.Valid(), tt.wantValid)
		})
	}
}

func TestType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		types     pouet.Types
		wantStr   string
		wantValid bool
	}{
		{"empty", pouet.Types{}, "", false},
		{"all invalid", pouet.Types{"fastdemo", "wild"}, "fastdemo, wild", false},
		{"mixed", pouet.Types{"demo", "fastdemo"}, "demo, fastdemo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, tt.types.Valid(), tt.wantValid)
			be.Equal(t, tt.types.String(), tt.wantStr)
		})
	}
}

func TestResponseGet(t *testing.T) {
	t.Parallel()
	r := pouet.Response{}
	ctx := t.Context()
	_, err := r.Get(ctx, 0)
	be.Err(t, err, pouet.ErrBadID)
	// this pings a remote server, so it is disabled.
	if testRemoteServers {
		_, err = r.Get(ctx, 1)
		be.Err(t, err, nil)
	}
}

func TestPouet(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	p := pouet.Production{}
	_, err := p.Get(ctx, 0)
	be.Err(t, err, pouet.ErrBadID)
	// this pings a remote server, so it is disabled.
	if testRemoteServers {
		_, err = p.Get(ctx, 1)
		be.Err(t, err, nil)
	}
}

func TestVotes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	v := pouet.Votes{}
	err := v.Votes(ctx, 0)
	be.Err(t, err, pouet.ErrBadID)
	// this pings a remote server, so it is disabled.
	if testRemoteServers {
		err = v.Votes(ctx, 1)
		be.Err(t, err, nil)
	}
}

func TestStars(t *testing.T) {
	t.Parallel()
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
		{"0 votes", args{0, 0, 0}, 0},
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
			t.Parallel()
			f := pouet.Stars(tt.args.up, tt.args.meh, tt.args.down)
			be.Equal(t, f, tt.want)
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

	type group = struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		pp := pouet.Production{}
		a, b := pp.Releasers()
		be.Equal(t, a, "")
		be.Equal(t, b, "")
		x, y, z := pp.Released()
		be.Equal(t, x, int16(0))
		be.Equal(t, y, int16(0))
		be.Equal(t, z, int16(0))
	})

	t.Run("one group and release date", func(t *testing.T) {
		t.Parallel()

		pp := pouet.Production{
			ReleaseDate: "1996-05-12",
			Groups:      []group{{ID: "1", Name: "Razor 1911"}},
		}
		a, b := pp.Releasers()
		be.Equal(t, a, "Razor 1911")
		be.Equal(t, b, "")
		x, y, z := pp.Released()
		be.Equal(t, x, int16(1996))
		be.Equal(t, y, int16(5))
		be.Equal(t, z, int16(12))
	})

	t.Run("multiple groups", func(t *testing.T) {
		t.Parallel()

		pp := pouet.Production{
			Groups: []group{
				{ID: "1", Name: "Future Crew"},
				{ID: "2", Name: "Triton"},
				{ID: "3", Name: "Ignored"},
			},
		}
		a, b := pp.Releasers()
		be.Equal(t, a, "Future Crew")
		be.Equal(t, b, "Triton")
	})
}

func TestPlatformType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prod     pouet.Production
		wantPlat tags.Tag
		wantSect tags.Tag
	}{
		{
			name:     "empty",
			prod:     pouet.Production{},
			wantPlat: tags.Tag(-1),
			wantSect: tags.Intro,
		},
		{
			name: "windows demo",
			prod: pouet.Production{
				Platforms: pouet.Platforms{Windows: pouet.Platform{Slug: "windows"}},
				Types:     pouet.Types{"demo"},
			},
			wantPlat: tags.Windows,
			wantSect: tags.Demo,
		},
		{
			name: "msdos artpack",
			prod: pouet.Production{
				Platforms: pouet.Platforms{MSDos: pouet.Platform{Slug: "msdos"}},
				Types:     pouet.Types{"artpack"},
			},
			wantPlat: tags.DOS,
			wantSect: tags.Pack,
		},
		{
			name: "dosgus bbstro",
			prod: pouet.Production{
				Platforms: pouet.Platforms{DosGus: pouet.Platform{Slug: "msdosgus"}},
				Types:     pouet.Types{"bbstro"},
			},
			wantPlat: tags.DOS,
			wantSect: tags.BBS,
		},
		{
			name: "diskmag",
			prod: pouet.Production{
				Platforms: pouet.Platforms{MSDos: pouet.Platform{Slug: "msdos"}},
				Types:     pouet.Types{"diskmag"},
			},
			wantPlat: tags.DOS,
			wantSect: tags.Mag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plat, sect := tt.prod.PlatformType()
			be.Equal(t, plat, tt.wantPlat)
			be.Equal(t, sect, tt.wantSect)
		})
	}
}
