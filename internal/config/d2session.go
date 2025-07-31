package config

import (
	"fmt"
	"log/slog"
)

type Hours int // Hours is a duration value

func (h Hours) LogValue() slog.Value {
	return slog.IntValue(int(h))
}

func (h Hours) String() string {
	if h == 1 {
		return "1 hour"
	}
	return fmt.Sprintf("%d hours", h)
}

func (h Hours) Int() int {
	return int(h)
}
