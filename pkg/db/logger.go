package db

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// 首先创建一个实现 gorm.Logger 接口的自定义 logger 结构体
type GormZapLogger struct {
    ZapLogger *zap.Logger
    LogLevel  logger.LogLevel
}

// 创建新的 logger 实例
func NewGormZapLogger(zapLogger *zap.Logger) *GormZapLogger {
    return &GormZapLogger{
        ZapLogger: zapLogger,
        LogLevel:  logger.Info, // 默认日志级别
    }
}

// 实现 gorm.Logger 接口的必要方法
func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
    newlogger := *l
    newlogger.LogLevel = level
    return &newlogger
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    l.ZapLogger.Sugar().Infof(msg, data...)
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    l.ZapLogger.Sugar().Warnf(msg, data...)
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    l.ZapLogger.Sugar().Errorf(msg, data...)
}

func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    // 将 SQL 查询信息记录到日志中
    fields := []zap.Field{
        zap.Duration("elapsed", elapsed),
        zap.String("sql", sql),
        zap.Int64("rows", rows),
    }
    
    if err != nil {
        fields = append(fields, zap.Error(err))
        l.ZapLogger.Error("gorm-trace", fields...)
        return
    }

    // 根据查询耗时决定日志级别
    if elapsed > time.Second {
        l.ZapLogger.Warn("gorm-trace-slow", fields...)
        return
    }

    l.ZapLogger.Debug("gorm-trace", fields...)
} 