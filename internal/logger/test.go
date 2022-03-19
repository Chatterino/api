package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewTest() Logger {
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := zapConfig.Build()
	zap.RedirectStdLog(logger)
	return logger.Sugar()
}
