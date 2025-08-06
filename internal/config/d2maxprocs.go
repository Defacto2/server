package config

import (
	"fmt"
	"log/slog"
)

type Threads uint

func (t Threads) LogValue() slog.Value {
	return slog.Uint64Value(uint64(t))
}

func (t Threads) Help() string {
	if t == 0 {
		return "The application will use all available CPU threads"
	}
	return "The application will limit usage of the CPU"
}

func (t Threads) String() string {
	if t == 0 {
		return "are not set"
	}
	return fmt.Sprintf("%d CPU threads", t)
}
