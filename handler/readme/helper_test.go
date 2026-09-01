package readme_test

import (
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/nalgeon/be"
)

func TestAddBytes(t *testing.T) {
	t.Parallel()

	str := []byte("hello world")
	pre := []byte("prefix text")
	suf := []byte("suffix text")

	x := readme.AddPrefix(str, pre)
	wantN := len(str) + len(pre) + 2
	be.Equal(t, len(x), wantN)
	want := string(pre) + "\n\n" + string(str)
	be.Equal(t, string(x), want)

	x = readme.AddSuffix(str, suf)
	wantN = len(str) + len(pre) + 2
	be.Equal(t, len(x), wantN)
	want = string(str) + "\n\n" + string(suf)
	be.Equal(t, string(x), want)
}
