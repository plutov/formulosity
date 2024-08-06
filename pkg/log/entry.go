package log

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Entry struct {
	logger *zap.SugaredLogger
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
func (e *Entry) With(args ...interface{}) *Entry {
	return &Entry{
		logger: e.logger.With(args...),
	}
}

// WithError add error tag with error value to the log.
func (e *Entry) WithError(err error) *Entry {
	return &Entry{
		logger: e.logger.With(errorLabel, err),
	}
}

// WithTime add time tag with time value to the log.
func (e *Entry) WithTime(t time.Time) *Entry {
	return &Entry{
		logger: e.logger.With(timeLabel, t),
	}
}

// WithContext add context tag with context value to the log.
func (e *Entry) WithContext(c context.Context) *Entry {
	return &Entry{
		logger: e.logger.With(contextLabel, c),
	}
}

// Debug uses fmt.Sprint to construct and log a message.
func (e *Entry) Debug(args ...interface{}) {
	e.logger.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func (e *Entry) Info(args ...interface{}) {
	e.logger.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (e *Entry) Warn(args ...interface{}) {
	e.logger.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (e *Entry) Error(args ...interface{}) {
	e.logger.Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (e *Entry) DPanic(args ...interface{}) {
	e.logger.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (e *Entry) Panic(args ...interface{}) {
	e.logger.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (e *Entry) Fatal(args ...interface{}) {
	e.logger.Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (e *Entry) Debugf(template string, args ...interface{}) {
	e.logger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (e *Entry) Infof(template string, args ...interface{}) {
	e.logger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (e *Entry) Warnf(template string, args ...interface{}) {
	e.logger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (e *Entry) Errorf(template string, args ...interface{}) {
	e.logger.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (e *Entry) DPanicf(template string, args ...interface{}) {
	e.logger.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (e *Entry) Panicf(template string, args ...interface{}) {
	e.logger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (e *Entry) Fatalf(template string, args ...interface{}) {
	e.logger.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func (e *Entry) Debugw(msg string, keysAndValues ...interface{}) {
	e.logger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *Entry) Infow(msg string, keysAndValues ...interface{}) {
	e.logger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *Entry) Warnw(msg string, keysAndValues ...interface{}) {
	e.logger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *Entry) Errorw(msg string, keysAndValues ...interface{}) {
	e.logger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (e *Entry) DPanicw(msg string, keysAndValues ...interface{}) {
	e.logger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (e *Entry) Panicw(msg string, keysAndValues ...interface{}) {
	e.logger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (e *Entry) Fatalw(msg string, keysAndValues ...interface{}) {
	e.logger.Fatalw(msg, keysAndValues...)
}
