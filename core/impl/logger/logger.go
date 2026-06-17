// Package logger
package logger

import (
	"fmt"
	"os"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
	conf   core.IConfig
	fields []zap.Field
}

func NewLogger(config core.IConfig) (*Logger, error) {
	logger := &Logger{
		conf: config,
	}
	err := logger.init()
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	return logger, nil
}

func (l *Logger) Log(fields ...any) error {
	l.withFields().Info(fmt.Sprintf("%v", fields...))
	return nil
}

func (l *Logger) Debug(msg string, fields ...any) {
	l.withFields().Debug(msg, l.convertFields(fields...)...)
}

func (l *Logger) Info(msg string, fields ...any) {
	l.withFields().Info(msg, l.convertFields(fields...)...)
}

func (l *Logger) Warn(msg string, fields ...any) {
	l.withFields().Warn(msg, l.convertFields(fields...)...)
}

func (l *Logger) Error(msg string, fields ...any) {
	l.withFields().Error(msg, l.convertFields(fields...)...)
}

func (l *Logger) Fatal(msg string, fields ...any) {
	l.withFields().Fatal(msg, l.convertFields(fields...)...)
}

func (l *Logger) Panic(msg string, fields ...any) {
	l.withFields().Panic(msg, l.convertFields(fields...)...)
}

func (l *Logger) AddGlobalFields(fields ...any) {
	for _, field := range fields {
		if f, ok := field.(zap.Field); ok {
			l.fields = append(l.fields, f)
		}
	}
}

func (l *Logger) init() error {
	fileEncoder := getFileEncoder()
	stdoutEncoder := getStdoutEncoder()

	if l.conf == nil {
		return cerrs.New("logger config not found")
	}
	infoWriter, err := getInfoLogFileWriter(l.conf)
	if err != nil {
		return err
	}
	errWriter, err := getErrLogFileWriter(l.conf)
	if err != nil {
		return err
	}

	errLevelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zap.ErrorLevel
	})
	infoLevelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zap.ErrorLevel
	})

	infoCore := zapcore.NewCore(fileEncoder, infoWriter, infoLevelEnabler)
	errCore := zapcore.NewCore(fileEncoder, errWriter, errLevelEnabler)
	coreArr := []zapcore.Core{infoCore, errCore}
	mode, err := core.GetString(l.conf, "mode")
	if err != nil {
		return cerrs.Wrap(err)
	}
	if mode == "debug" {
		coreArr = append(coreArr, zapcore.NewCore(stdoutEncoder, getStdoutWriter(), zap.DebugLevel))
	}

	l.logger = zap.New(zapcore.NewTee(coreArr...), zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

func (l *Logger) withFields() *zap.Logger {
	if len(l.fields) > 0 {
		return l.logger.With(l.fields...)
	}
	return l.logger
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

func getLogFileConfig(conf core.IConfig, filename string) (*lumberjack.Logger, error) {
	filePath, err := core.GetString(conf, "logger.file_path")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	maxSize, err := core.GetInt(conf, "logger.max_size")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	maxAge, err := core.GetInt(conf, "logger.max_age")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	maxBackups, err := core.GetInt(conf, "logger.max_backups")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	return &lumberjack.Logger{
		Filename:   filePath + "/" + filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   false,
	}, nil
}

func getErrLogFileWriter(conf core.IConfig) (zapcore.WriteSyncer, error) {
	lumberJackLogger, err := getLogFileConfig(conf, "error.log")
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(lumberJackLogger), nil
}

func getInfoLogFileWriter(conf core.IConfig) (zapcore.WriteSyncer, error) {
	lumberJackLogger, err := getLogFileConfig(conf, "info.log")
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(lumberJackLogger), nil
}

func getStdoutWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}
