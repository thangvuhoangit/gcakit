// Package log configures a new logger for an application.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"
)

var logLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLogLevel(level string) zapcore.Level {
	l, exist := logLevelMap[level]
	if !exist {
		return zapcore.DebugLevel
	}

	return l
}

// NewLogger creates a new logger.
func NewZapLogger(config Config) logur.LoggerFacade {
	logWriter := zapcore.AddSync(os.Stdout)

	var encoderCfg zapcore.EncoderConfig
	var encoder zapcore.Encoder

	if config.Mode == "development" {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	encoderCfg.NameKey = "[SERVICE]"
	encoderCfg.TimeKey = "[TIME]"
	encoderCfg.LevelKey = "[LEVEL]"
	encoderCfg.FunctionKey = "[CALLER]"
	encoderCfg.CallerKey = "[LINE]"
	encoderCfg.MessageKey = "[MESSAGE]"
	encoderCfg.StacktraceKey = "[STACKTRACE]"

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.EncodeName = zapcore.FullNameEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	if !config.NoColor {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoder = zapcore.NewConsoleEncoder(encoderCfg)

	switch config.Format {
	case "logfmt":
		// Already the default

	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(getLogLevel(config.Level)))
	logger := zap.New(core)

	if config.Mode == "development" {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	return zapadapter.New(logger)
}
