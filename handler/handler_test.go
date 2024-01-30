package handler_test

import (
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	tr, err := c.Registry()
	assert.Nil(t, tr)
	assert.Error(t, handler.ErrTmpl, err)
}
