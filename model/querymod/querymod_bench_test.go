package querymod_test

import (
	"testing"

	"github.com/Defacto2/server/model/querymod"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkSAdvert(b *testing.B) {
	for b.Loop() {
		_ = querymod.SAdvert()
	}
}

func BenchmarkPWindows(b *testing.B) {
	for b.Loop() {
		_ = querymod.PWindows()
	}
}

func BenchmarkPDos(b *testing.B) {
	for b.Loop() {
		_ = querymod.PDos()
	}
}

func BenchmarkAdvertExpr(b *testing.B) {
	for b.Loop() {
		_ = querymod.AdvertExpr()
	}
}

func BenchmarkDOSExpr(b *testing.B) {
	for b.Loop() {
		_ = querymod.DOSExpr()
	}
}

func BenchmarkAnsiBBSExpr(b *testing.B) {
	for b.Loop() {
		_ = querymod.AnsiBBSExpr()
	}
}

func BenchmarkWindowsPackExpr(b *testing.B) {
	for b.Loop() {
		_ = querymod.WindowsPackExpr()
	}
}
