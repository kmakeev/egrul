package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config - конфигурация логгера
type Config struct {
	Level       string // debug, info, warn, error
	Format      string // json, console
	ServiceName string
}

// Context keys для trace_id, span_id, request_id
type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	SpanIDKey    contextKey = "span_id"
	RequestIDKey contextKey = "request_id"
)

// NewLogger - создание централизованного Zap логгера
func NewLogger(cfg Config) (*zap.Logger, error) {
	var zapCfg zap.Config

	// Выбор уровня логирования
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// Конфигурация в зависимости от формата
	if cfg.Format == "json" {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.TimeKey = "timestamp"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Создание логгера
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	// Добавление service field
	logger = logger.With(zap.String("service", cfg.ServiceName))

	return logger, nil
}

// WithTraceContext - добавляет trace_id, span_id, request_id из контекста в logger
func WithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	fields := []zap.Field{}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if spanID, ok := ctx.Value(SpanIDKey).(string); ok && spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if len(fields) > 0 {
		return logger.With(fields...)
	}

	return logger
}

// Context helper functions
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func ContextWithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}
