package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormZapLogger is a custom logger for GORM that uses Zap.
type GormZapLogger struct {
	ZapLogger *zap.Logger
	LogLevel  logger.LogLevel
}

// NewGormZapLogger creates a new GormZapLogger.
func NewGormZapLogger(zapLogger *zap.Logger) *GormZapLogger {
	return &GormZapLogger{
		ZapLogger: zapLogger,
		LogLevel:  logger.Info, // Default log level
	}
}

// LogMode sets the log level.
func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs informational messages.
func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn logs warning messages.
func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error logs error messages.
func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace logs SQL queries and their execution details.
func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	// Log errors, but ignore "record not found" as it's a normal GORM behavior
	case err != nil && l.LogLevel >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.ZapLogger.Error("GORM Trace",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	// Log slow queries at the WARN level
	case elapsed > 200*time.Millisecond && l.LogLevel >= logger.Warn:
		l.ZapLogger.Warn("GORM Trace [SLOW]",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	// Log all other queries at the INFO level
	case l.LogLevel >= logger.Info:
		l.ZapLogger.Info("GORM Trace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
