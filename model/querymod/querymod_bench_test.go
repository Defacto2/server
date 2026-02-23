package querymod_test

import (
	"testing"

	"github.com/Defacto2/server/model/querymod"
)

// Benchmarks for tag functions with caching optimization.

func BenchmarkSAdvert(b *testing.B) {
	for range b.N {
		_ = querymod.SAdvert()
	}
}

func BenchmarkPWindows(b *testing.B) {
	for range b.N {
		_ = querymod.PWindows()
	}
}

func BenchmarkPDos(b *testing.B) {
	for range b.N {
		_ = querymod.PDos()
	}
}

func BenchmarkAdvertExpr(b *testing.B) {
	for range b.N {
		_ = querymod.AdvertExpr()
	}
}

func BenchmarkDOSExpr(b *testing.B) {
	for range b.N {
		_ = querymod.DOSExpr()
	}
}

func BenchmarkAnsiBBSExpr(b *testing.B) {
	for range b.N {
		_ = querymod.AnsiBBSExpr()
	}
}

func BenchmarkWindowsPackExpr(b *testing.B) {
	for range b.N {
		_ = querymod.WindowsPackExpr()
	}
}
