package logging

import (
	"context"
	"log"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CtxField is the type used for storing zap fields in a context.Context.
type CtxField string

const zapFieldsKey = CtxField("zapFields")

// ZapFields represents a set of zap fields keyed by their names.
type ZapFields map[string]zap.Field

// Append returns a new ZapFields map containing the receiver's fields merged
// with the provided fields. If a key already exists it will be overwritten by
// the latest occurrence.
func (zf ZapFields) Append(fields ...zap.Field) ZapFields {
	zfCopy := make(ZapFields)
	for k, v := range zf {
		zfCopy[k] = v
	}

	for _, f := range fields {
		zfCopy[f.Key] = f
	}

	return zfCopy
}

// ZapLogger wraps a zap.Logger and provides helpers for working with context
// aware structured logging.
type ZapLogger struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

// NewNopLogger creates a ZapLogger that discards all log output.
func NewNopLogger() *ZapLogger {
	return &ZapLogger{
		logger: zap.NewNop(),
		level:  zap.NewAtomicLevel(),
	}
}

// NewZapLogger creates a ZapLogger configured with the provided log level.
func NewZapLogger(level zapcore.Level) (*ZapLogger, error) {
	atomic := zap.NewAtomicLevelAt(level)
	settings := defaultSettings(atomic)

	l, err := settings.config.Build(settings.opts...)
	if err != nil {
		return nil, err //nolint:wrapcheck // unnecessary
	}

	return &ZapLogger{
		logger: l,
		level:  atomic,
	}, nil
}

// WithContextFields stores zap fields in the supplied context and returns the
// derived context containing them.
func WithContextFields(ctx context.Context, fields ...zap.Field) context.Context {
	ctxFields, _ := ctx.Value(zapFieldsKey).(ZapFields)
	if ctxFields == nil {
		ctxFields = make(ZapFields)
	}

	merged := ctxFields.Append(fields...)
	return context.WithValue(ctx, zapFieldsKey, merged)
}

func maskField(f zap.Field) zap.Field {
	if f.Key == "password" {
		return zap.String(f.Key, "******")
	}

	if f.Key == "email" {
		email := f.String
		parts := strings.Split(email, "@")
		if len(parts) == 2 { //nolint:gomnd // unnecessary
			return zap.String(f.Key, "***@"+parts[1])
		}
	}
	return f
}

// Sync flushes any buffered log entries.
func (z *ZapLogger) Sync() {
	_ = z.logger.Sync()
}

func withCtxFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	fs := make(ZapFields)

	ctxFields, ok := ctx.Value(zapFieldsKey).(ZapFields)
	if ok {
		fs = ctxFields
	}

	fs = fs.Append(fields...)

	maskedFields := make([]zap.Field, 0, len(fs))
	for _, f := range fs {
		maskedFields = append(maskedFields, maskField(f))
	}

	return maskedFields
}

// InfoCtx logs an info level message with fields from the provided context.
func (z *ZapLogger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Info(msg, withCtxFields(ctx, fields...)...)
}

// DebugCtx logs a debug level message with fields from the provided context.
func (z *ZapLogger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Debug(msg, withCtxFields(ctx, fields...)...)
}

// WarnCtx logs a warning level message with fields from the provided context.
func (z *ZapLogger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Warn(msg, withCtxFields(ctx, fields...)...)
}

// ErrorCtx logs an error level message with fields from the provided context.
func (z *ZapLogger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Error(msg, withCtxFields(ctx, fields...)...)
}

// FatalCtx logs a fatal level message with fields from the provided context
// and then exits the application.
func (z *ZapLogger) FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, withCtxFields(ctx, fields...)...)
}

// PanicCtx logs a panic level message with fields from the provided context
// and then panics.
func (z *ZapLogger) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Panic(msg, withCtxFields(ctx, fields...)...)
}

// SetLevel updates the logging level of the underlying logger.
func (z *ZapLogger) SetLevel(level zapcore.Level) {
	z.level.SetLevel(level)
}

// Std returns a standard library logger that writes through this ZapLogger.
func (z *ZapLogger) Std() *log.Logger {
	return zap.NewStdLog(z.logger)
}

// Raw exposes the underlying zap.Logger instance.
func (z *ZapLogger) Raw() *zap.Logger {
	return z.logger
}
