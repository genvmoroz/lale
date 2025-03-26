// TODO: remove this package and use another logger package
package logger

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// todo: use zap logger

type ctxLoggerKey struct{}

var key ctxLoggerKey

func ContextWithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, key, entry)
}

func FromContext(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(key).(*logrus.Entry); ok {
		return logger
	}
	return WithCorrelationID()
}

func WithCorrelationID() *logrus.Entry {
	return logrus.
		StandardLogger().
		WithFields(logrus.Fields{"CorrelationID": uuid.NewString()})
}
