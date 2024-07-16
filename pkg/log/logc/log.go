package logc

import (
	"context"
	"github.com/peckfly/gopeck/pkg/log"
	"go.uber.org/zap"
)

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	log.Context(ctx).Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	log.Context(ctx).Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	log.Context(ctx).Error(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	log.Context(ctx).Fatal(msg, fields...)
}
