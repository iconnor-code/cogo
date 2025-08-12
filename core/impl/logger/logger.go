// Package logger
package logger

import (
	"os"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
	conf   map[string]any
	fields []zap.Field
	mode   string
}

func LoggerModeOption(mode string) core.LoggerOption {
	return func(l core.ILogger) error {
		l.(*Logger).mode = mode
		return nil
	}
}

func NewLogger(config core.IConfig, opts ...core.LoggerOption) (*Logger, error) {
	logger := &Logger{
		conf: config.Get("logger").(map[string]any),
	}
	for _, opt := range opts {
		err := opt(logger)
		if err != nil {
			return nil, err
		}
	}
	err := logger.init()
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	return logger, nil
}

func (l *Logger) Debug(msg string, fields ...any) {
	l.withFields()
	l.logger.Debug(msg, l.convertFields(fields...)...)
}

func (l *Logger) Info(msg string, fields ...any) {
	l.withFields()
	l.logger.Info(msg, l.convertFields(fields...)...)
}

func (l *Logger) Warn(msg string, fields ...any) {
	l.withFields()
	l.logger.Warn(msg, l.convertFields(fields...)...)
}

func (l *Logger) Error(msg string, fields ...any) {
	l.withFields()
	l.logger.Error(msg, l.convertFields(fields...)...)
}

func (l *Logger) Fatal(msg string, fields ...any) {
	l.withFields()
	l.logger.Fatal(msg, l.convertFields(fields...)...)
}

func (l *Logger) Panic(msg string, fields ...any) {
	l.withFields()
	l.logger.Panic(msg, l.convertFields(fields...)...)
}

func (l *Logger) AddGlobalFields(fields ...any) {
	for _, field := range fields {
		l.fields = append(l.fields, field.(zap.Field))
	}
}

func (l *Logger) init() error {
	fileEncoder := getFileEncoder()
	stdoutEncoder := getStdoutEncoder()

	if l.conf == nil {
		return cerrs.New("logger config not found")
	}
	infoWriter := getInfoLogFileWriter(l.conf)
	errWriter := getErrLogFileWriter(l.conf)

	errLevelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zap.ErrorLevel
	})
	infoLevelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zap.ErrorLevel
	})

	infoCore := zapcore.NewCore(fileEncoder, infoWriter, infoLevelEnabler)
	errCore := zapcore.NewCore(fileEncoder, errWriter, errLevelEnabler)
	coreArr := []zapcore.Core{infoCore, errCore}
	if l.mode == "debug" {
		coreArr = append(coreArr, zapcore.NewCore(stdoutEncoder, getStdoutWriter(), zap.DebugLevel))
	}

	l.logger = zap.New(zapcore.NewTee(coreArr...), zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

func (l *Logger) withFields() error {
	for _, field := range l.fields {
		l.logger.With(field)
	}
	return nil
}

func (l *Logger) convertFields(fields ...any) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		if f, ok := field.(zap.Field); ok {
			zapFields = append(zapFields, f)
		} else if e, ok := field.(error); ok {
			zapFields = append(zapFields, zap.Error(e))
		} else {
			zapFields = append(zapFields, zap.Any("field", field))
		}
	}
	return zapFields
}

func getFileEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.FullCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

func getStdoutEncoder() zapcore.Encoder {
	encodeConfig := zap.NewDevelopmentEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encodeConfig.EncodeCaller = zapcore.FullCallerEncoder
	return zapcore.NewConsoleEncoder(encodeConfig)
}

func getErrLogFileWriter(conf map[string]any) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf["file_path"].(string) + "/error.log",
		MaxSize:    conf["max_size"].(int),
		MaxAge:     conf["max_age"].(int),
		MaxBackups: conf["max_backups"].(int),
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func getInfoLogFileWriter(conf map[string]any) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf["file_path"].(string) + "/info.log", // 文件位置
		MaxSize:    conf["max_size"].(int),                   // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     conf["max_age"].(int),                    // 保留旧文件的最大天数
		MaxBackups: conf["max_backups"].(int),                // 保留旧文件的最大个数
		Compress:   false,                                    // 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}

func getStdoutWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}
