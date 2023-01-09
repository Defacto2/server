package logger

import (
	"fmt"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Initialize() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan-02-15:04:05.00")

	logger, err := config.Build()
	if err != nil {
		log.Fatalln(fmt.Errorf("could not initialize the zap logger: %w", err))
	}
	return logger
}
