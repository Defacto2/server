package htmx_test

import (
	"testing"

	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

const (
	unid = "123e4567-e89b-12d3-a456-426614174000"
	unv4 = "bb2310e1-93aa-475e-8b88-59eb1fb984a4"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want error
	}{
		{
			name: "absolute path",
			path: "/absolute/path",
			want: htmx.ErrPath,
		},
		{
			name: "clean path",
			path: "relative/path",
			want: nil,
		},
		{
			name: "clean path",
			path: "relative/path/",
			want: nil,
		},
		{
			name: "unclean path 1",
			path: "relative/../path",
			want: htmx.ErrPath,
		},
		{
			name: "unclean path 2",
			path: "./relative/path",
			want: htmx.ErrPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := htmx.Validate(tt.path)
			be.Err(t, err, tt.want)
		})
	}
}

func TestPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		unid     string
		path     string
		wantUnid string
		wantName string
		wantErr  bool
	}{
		{
			name:     "valid unid and path",
			unid:     unid,
			path:     "relative/path",
			wantUnid: unid,
			wantName: "relative/path",
			wantErr:  false,
		},
		{
			name:     "invalid unid",
			unid:     "invalid-unid",
			path:     "relative/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
		{
			name:     "invalid path",
			unid:     unid,
			path:     "/absolute/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := testutil.NewContext(t, "")
			c.SetPathValues(echo.PathValues{
				{Name: "unid", Value: tt.unid},
				{Name: "path", Value: tt.path},
			})

			gotUnid, gotName, err := htmx.Path(c)
			got := (err != nil)
			be.Equal(t, got, tt.wantErr)
			be.Equal(t, tt.wantUnid, gotUnid)
			be.Equal(t, tt.wantName, gotName)
		})
	}
}

func TestUUID(t *testing.T) {
	t.Parallel()

	c := testutil.NewForm(t, "", "unid", unid)
	_, err := htmx.UUID(c)
	be.Err(t, err)
}
