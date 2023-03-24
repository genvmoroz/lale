package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

var key = struct {
	correlationID string
}{
	correlationID: "3ef61cf6-ffe1-4ac3-87f2-da400fc71e6f",
}

func ContextWithLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

// FromContext returns *logrus.Logger stored in ctx,
// if there is no logger stored it returns default logger
func FromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(key).(*logrus.Logger); ok {
		return logger
	}
	return logrus.StandardLogger()
}
