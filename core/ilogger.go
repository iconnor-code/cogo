package core

type ILogger interface {
	Log(...any) error
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	Panic(msg string, fields ...any)
	AddGlobalFields(fields ...any)
}

type LoggerOption func(l ILogger) error
