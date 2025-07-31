package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes and returns a new zap logger.
func Init(projectRoot string) (*zap.Logger, error) {
	// Base encoder configuration for file logs (JSON format)
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		TimeKey:      "time",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	logDir := filepath.Join(projectRoot, "logs")

	// Create a core for each level, which writes ONLY that level to a file.
	debugFileCore, err := newFileCore(logDir, zapcore.DebugLevel, encoderConfig)
	if err != nil {
		return nil, err
	}
	infoFileCore, err := newFileCore(logDir, zapcore.InfoLevel, encoderConfig)
	if err != nil {
		return nil, err
	}
	warnFileCore, err := newFileCore(logDir, zapcore.WarnLevel, encoderConfig)
	if err != nil {
		return nil, err
	}
	errorFileCore, err := newFileCore(logDir, zapcore.ErrorLevel, encoderConfig)
	if err != nil {
		return nil, err
	}

	// Create a separate core for the console with a more readable format.
	consoleCore := newConsoleCore()

	// Combine all cores. A log entry will be sent to all of them,
	// and each will decide whether to write it based on its LevelEnabler.
	core := zapcore.NewTee(
		debugFileCore,
		infoFileCore,
		warnFileCore,
		errorFileCore,
		consoleCore,
	)

	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}

// newFileCore creates a core that writes a specific log level to a rotating file.
func newFileCore(logDir string, level zapcore.Level, encoderConfig zapcore.EncoderConfig) (zapcore.Core, error) {

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create log directory: %w", err)
	}

	// Create a log file for each level, named like '2025-07-30-info.log'
	fileName := filepath.Join(logDir, fmt.Sprintf("%s-%s.log", time.Now().Format("2006-01-02"), level.String()))

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   true,
	})

	// This LevelEnablerFunc is the key to splitting logs. It ensures
	// that this core only handles logs of the exact specified level.
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l == level
	})

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writer,
		levelEnabler,
	), nil
}

// newConsoleCore creates a core that writes to the console.
func newConsoleCore() zapcore.Core {
	// Log everything from Debug and above to the console.
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.DebugLevel
	})

	// Use a more human-readable encoder for the console.
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color to levels

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		levelEnabler,
	)
}
