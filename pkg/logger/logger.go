package logger

import (
	"context"
	"os"

	"github.com/iconnor-code/cogo/pkg/config"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

type Logger struct {
	logger *zap.Logger
	conf   *config.Conf
	ctx    context.Context
	opts   []zap.Option
	fieldKeys []contextKey
}

func NewLogger(conf *config.Conf) *Logger {
	logger := &Logger{
		conf: conf,
	}
	logger.opts = []zap.Option{
		zap.AddCaller(),
		zap.Hooks(logger.fieldHook),
	}
	logger.initZapLogger()
	return logger
}

func (l *Logger) SetContext(ctx context.Context) *Logger {
	l.ctx = ctx
	return l
}

func (l *Logger) WithField(key contextKey, value interface{}) *Logger {
	l.fieldKeys = append(l.fieldKeys, key)
	l.ctx = context.WithValue(l.ctx, key, value)
	return l
}

func (l *Logger) Log() *zap.Logger {
	return l.logger
}

func (l *Logger) fieldHook(entry zapcore.Entry) error {
	for _, key := range l.fieldKeys {
		value := l.ctx.Value(key)
		if value != nil {
			l.logger.With(zap.Any(string(key), value))
		}
	}
	return nil
}

func (l *Logger) initZapLogger() {
	fileEncoder := l.getFileEncoder()
	stdoutEncoder := l.getStdoutEncoder()

	infoWriter := l.getInfoLogFileWriter()
	errWriter := l.getErrLogFileWriter()

	errLevelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zap.ErrorLevel
	})
	infoLevelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zap.ErrorLevel
	})

	infoCore := zapcore.NewCore(fileEncoder, infoWriter, infoLevelEnabler)
	errCore := zapcore.NewCore(fileEncoder, errWriter, errLevelEnabler)
	coreArr := []zapcore.Core{infoCore, errCore}
	if l.conf.Mode == "debug" {
		coreArr = append(coreArr, zapcore.NewCore(stdoutEncoder, l.getStdoutWriter(), zap.DebugLevel))
	}

	logger := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller())

	l.logger = logger
}

func (l *Logger) getFileEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.FullCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

func (l *Logger) getStdoutEncoder() zapcore.Encoder {
	encodeConfig := zap.NewDevelopmentEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encodeConfig.EncodeCaller = zapcore.FullCallerEncoder
	return zapcore.NewConsoleEncoder(encodeConfig)
}

func (l *Logger) getErrLogFileWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   l.conf.Log.FilePath + "/error.log",
		MaxSize:    l.conf.Log.MaxSize,
		MaxAge:     l.conf.Log.MaxAge,
		MaxBackups: l.conf.Log.MaxBackups,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func (l *Logger) getInfoLogFileWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   l.conf.Log.FilePath + "/info.log", // 文件位置
		MaxSize:    l.conf.Log.MaxSize,                // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     l.conf.Log.MaxAge,                 // 保留旧文件的最大天数
		MaxBackups: l.conf.Log.MaxBackups,             // 保留旧文件的最大个数
		Compress:   false,                             // 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}

func (l *Logger) getStdoutWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}
