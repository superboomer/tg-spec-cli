package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	if level == "silent" {
		return zap.New(zapcore.NewNopCore()), nil
	}

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}

	// Use console encoding for more human-readable logs
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: true, // Enable development mode for friendlier output
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console", // <-- human-friendly
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "",
			CallerKey:      "",
			MessageKey:     "msg",
			StacktraceKey:  "",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,             // [DEBUG] [INFO] [ERROR]
			EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05"), // Short time
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()
}
