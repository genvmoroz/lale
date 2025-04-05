// TODO: remove this package and use another logger package
package logger

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// todo: use zap logger

type ctxLoggerKey struct{}

// getLoggerKey returns a new instance of ctxLoggerKey.
// This is used as a context key for storing logger entries in the context.
func getLoggerKey() ctxLoggerKey {
	return ctxLoggerKey{}
}

func ContextWithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, getLoggerKey(), entry)
}

func FromContext(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(getLoggerKey()).(*logrus.Entry); ok {
		return logger
	}
	return WithCorrelationID()
}

func WithCorrelationID() *logrus.Entry {
	return logrus.
		StandardLogger().
		WithFields(logrus.Fields{"CorrelationID": uuid.NewString()})
}
