package log

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEncodingFormat string

const (
	errorLabel   = "error"
	contextLabel = "context"
	timeLabel    = "usertime"

	EncodingConsoleFormat LogEncodingFormat = "console"
	EncodingJsonFormat    LogEncodingFormat = "json"
)

var (
	config     zap.Config
	loggerName string
)

// init creates logger with default configuration
func init() {
	config = getDefaultLoggerConfig()
	logger, _ := config.Build()
	initLoggerWithConfig(logger)
}

// initLoggerWithConfig with initial needed config all the time after creating the instance
func initLoggerWithConfig(newLogger *zap.Logger) {
	logger := newLogger.Named(loggerName)
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)
}

func getLoggerContext() *zap.SugaredLogger {
	return zap.S()
}

// getDefaultEncodeConfig returns default encoding config
func getDefaultEncodeConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// getDefaultLoggerConfig returns default logger configuration
func getDefaultLoggerConfig() zap.Config {
	return zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    getDefaultEncodeConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
}

// SetLogLevel creates new logger with new log level.
// If logLevel is not valid, then default log level will be set to 'debug'.
func SetLogLevel(logLevel string) {
	switch logLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "dpanic":
		config.Level = zap.NewAtomicLevelAt(zap.DPanicLevel)
	case "panic":
		config.Level = zap.NewAtomicLevelAt(zap.PanicLevel)
	case "fatal":
		config.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapLogger, _ := config.Build()
	initLoggerWithConfig(zapLogger)
}

// SetEncoding sets encoding for the log, valid formats are json and console.
func SetEncoding(encodingFormat LogEncodingFormat) {
	config.Encoding = string(encodingFormat)
	zapLogger, _ := config.Build()
	initLoggerWithConfig(zapLogger)
}

// Named adds a sub-scope to the logger's name. See Logger.Named for details.
// Set it to the service name to find the logs related to service easily.
func Named(name string) {
	loggerName = name
	zapLogger := zap.S().Desugar()
	zap.ReplaceGlobals(zapLogger)
	initLoggerWithConfig(zapLogger)
}

// With adds a variadic number of fields to the logging context. It accepts a
// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
// processing pairs, the first element of the pair is used as the field key
// and the second as the field value.
//
// For example,
//
//	 sugaredLogger.With(
//	   "hello", "world",
//	   "failure", errors.New("oh no"),
//	   Stack(),
//	   "count", 42,
//	   "user", User{Name: "alice"},
//	)
//
// is the equivalent of
//
//	unsugared.With(
//	  String("hello", "world"),
//	  String("failure", "oh no"),
//	  Stack(),
//	  Int("count", 42),
//	  Object("user", User{Name: "alice"}),
//	)
//
// Note that the keys in key-value pairs should be strings. In development,
// passing a non-string key panics. In production, the logger is more
// forgiving: a separate error is logged, but the key-value pair is skipped
// and execution continues. Passing an orphaned key triggers similar behavior:
// panics in development and errors in production.
func With(args ...interface{}) *Entry {
	return &Entry{
		logger: getLoggerContext().With(args...),
	}
}

// WithError add error tag with error value to the log.
func WithError(err error) *Entry {
	return &Entry{
		logger: getLoggerContext().With(errorLabel, err),
	}
}

// WithTime add time tag with time value to the log.
func WithTime(t time.Time) *Entry {
	return &Entry{
		logger: getLoggerContext().With(timeLabel, t),
	}
}

// WithContext add context tag with context value to the log.
func WithContext(c context.Context) *Entry {
	return &Entry{
		logger: getLoggerContext().With(contextLabel, c),
	}
}

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	getLoggerContext().Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	getLoggerContext().Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	getLoggerContext().Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	getLoggerContext().Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanic(args ...interface{}) {
	getLoggerContext().DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	getLoggerContext().Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	getLoggerContext().Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	getLoggerContext().Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	getLoggerContext().Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	getLoggerContext().Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	getLoggerContext().Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicf(template string, args ...interface{}) {
	getLoggerContext().DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	getLoggerContext().Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	getLoggerContext().Fatalf(template, args...)
}
