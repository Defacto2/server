package html3_test

import (
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/stretchr/testify/assert"
)

func TestClauses(t *testing.T) {
	tests := []string{
		html3.NameAsc,
		html3.NameDes,
		html3.PublAsc,
		html3.PublDes,
		html3.PostAsc,
		html3.PostDes,
		html3.SizeAsc,
		html3.SizeDes,
		html3.DescAsc,
		html3.DescDes,
	}
	for i, s := range tests {
		assert.Equal(t, int(html3.Clauses(s)), i)
	}
	assert.Equal(t, int(html3.Clauses("")),
		int(html3.Clauses(html3.NameAsc)), "default should be name asc")
}
