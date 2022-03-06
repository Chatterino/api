package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewTest() Logger {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := zapConfig.Build()
	zap.RedirectStdLog(logger)
	return logger.Sugar()
}
