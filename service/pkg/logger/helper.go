package logger

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var key = struct {
	correlationID string
}{
	correlationID: "3ef61cf6-ffe1-4ac3-87f2-da400fc71e6f",
}

func ContextWithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, key, entry)
}

func FromContext(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(key).(*logrus.Entry); ok {
		return logger
	}
	return logrus.
		StandardLogger().
		WithFields(logrus.Fields{"ID": uuid.NewString()})
}
