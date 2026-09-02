package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkSumm01(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = model.SummCols()
		}
	})
}

func BenchmarkSumm02(b *testing.B) {
	model.SummCols()
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = model.SummCols()
		}
	})
}

func BenchmarkSumm03(b *testing.B) {
	model.SummCols()
	b.Run("", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = model.SummCols()
			}
		})
	})
}
