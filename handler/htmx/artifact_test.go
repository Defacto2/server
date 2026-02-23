package htmx_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/Defacto2/server/handler/htmx"
	"github.com/nalgeon/be"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "absolute path",
			path:    "/absolute/path",
			wantErr: htmx.ErrPath,
		},
		{
			name:    "clean path",
			path:    "relative/path",
			wantErr: nil,
		},
		{
			name:    "clean path",
			path:    "relative/path/",
			wantErr: nil,
		},
		{
			name:    "unclean path 1",
			path:    "relative/../path",
			wantErr: htmx.ErrPath,
		},
		{
			name:    "unclean path 2",
			path:    "./relative/path",
			wantErr: htmx.ErrPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := htmx.Validate(tt.path)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.wantErr != nil {
				t.Errorf("Validate() expected error = %v, got nil", tt.wantErr)
			}
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
			unid:     "123e4567-e89b-12d3-a456-426614174000",
			path:     "relative/path",
			wantUnid: "123e4567-e89b-12d3-a456-426614174000",
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
			unid:     "123e4567-e89b-12d3-a456-426614174000",
			path:     "/absolute/path",
			wantUnid: "",
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newContext()
			c.SetParamNames("unid", "path")
			c.SetParamValues(tt.unid, url.QueryEscape(tt.path))

			gotUnid, gotName, err := htmx.Path(c)
			got := (err != nil)
			be.Equal(t, got, tt.wantErr)
			be.Equal(t, tt.wantUnid, gotUnid)
			be.Equal(t, tt.wantName, gotName)
		})
	}
}
