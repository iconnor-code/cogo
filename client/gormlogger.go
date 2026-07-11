package client

import (
	"context"
	"fmt"
	"time"

	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormZapLogger 首先创建一个实现 gorm.Logger 接口的自定义 logger 结构体
type GormZapLogger struct {
	logger core.ILogger
	level  logger.LogLevel
}

// NewGormZapLogger 创建新的 logger 实例
func NewGormZapLogger(logger core.ILogger) *GormZapLogger {
	return &GormZapLogger{
		logger: logger,
		// LogLevel: logger.Info, // 默认日志级别
	}
}

// LogMode 实现 gorm.Logger 接口的必要方法
func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.level = level
	return &newlogger
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...any) {
	l.logger.Info(fmt.Sprintf(msg, data...))
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...any) {
	l.logger.Error(fmt.Sprintf(msg, data...))
}

// ParamsFilter keeps SQL structure observable without writing query parameters
// such as passwords, verification codes, and tokens to logs.
func (l *GormZapLogger) ParamsFilter(_ context.Context, sql string, _ ...any) (string, []any) {
	return sql, nil
}

func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 将 SQL 查询信息记录到日志中
	fields := []any{
		zap.Duration("elapsed", elapsed),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.logger.Error("gorm-trace", fields...)
		return
	}

	// 根据查询耗时决定日志级别
	if elapsed > time.Second {
		l.logger.Warn("gorm-trace-slow", fields...)
		return
	}

	l.logger.Debug("gorm-trace", fields...)
}
