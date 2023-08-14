package helper

// Package time.go contains functions that return a time.Time value.

import (
	"math/big"
	"time"
)

// Latency returns the stored, current local time.
func Latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}
