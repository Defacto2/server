package postgres_test

import (
	"testing"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestConnection_URL(t *testing.T) {
	type fields struct {
		Protocol  string
		User      string
		Password  string
		HostName  string
		HostPort  int
		Database  string
		NoSSLMode bool
	}
	tests := []struct {
		fields fields
		want   string
	}{
		{fields{}, "postgres://localhost:5432"},
		{fields{Protocol: "dbproto"}, "dbproto://localhost:5432"},
		{fields{User: "dbuser"}, "postgres://dbuser@localhost:5432"},
		{fields{User: "dbuser", Password: "xyz"}, "postgres://dbuser:xyz@localhost:5432"},
		{fields{HostName: "myserver"}, "postgres://myserver:5432"},
		{fields{HostName: "myserver", HostPort: 5678}, "postgres://myserver:5678"},
		{
			fields{HostName: "myserver", HostPort: 5678, NoSSLMode: true},
			"postgres://myserver:5678?sslmode=disable",
		},
		{
			fields{Database: "my_db", HostName: "myserver", HostPort: 5678, NoSSLMode: true},
			"postgres://myserver:5678/my_db?sslmode=disable",
		},
	}
	for _, tt := range tests {
		c := postgres.Connection{
			Protocol:  tt.fields.Protocol,
			User:      tt.fields.User,
			Password:  tt.fields.Password,
			HostName:  tt.fields.HostName,
			HostPort:  tt.fields.HostPort,
			Database:  tt.fields.Database,
			NoSSLMode: tt.fields.NoSSLMode,
		}
		assert := assert.New(t)
		assert.Equal(tt.want, c.URL(), "url string does not match")
	}
}
