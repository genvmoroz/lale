package logger_test

import (
	"testing"

	"github.com/genvmoroz/lale/service/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestContextWithLogger(t *testing.T) {
	t.Run("should store logger in context", func(t *testing.T) {
		// Given
		ctx := t.Context()
		entry := logrus.NewEntry(logrus.New())
		entry = entry.WithField("CorrelationID", "123")

		// When
		ctxWithLogger := logger.ContextWithLogger(ctx, entry)

		// Then
		retrievedLogger := logger.FromContext(ctxWithLogger)
		require.Equal(t, entry, retrievedLogger)
	})

	t.Run("should not modify original context", func(t *testing.T) {
		// Given
		ctx := t.Context()
		entry := logrus.NewEntry(logrus.New())
		entry = entry.WithField("CorrelationID", "123")

		// When
		_ = logger.ContextWithLogger(ctx, entry)

		// Then
		retrievedLogger := logger.FromContext(ctx)
		require.NotEqual(t, entry, retrievedLogger)
	})
}

func TestFromContext(t *testing.T) {
	t.Run("should return logger from context", func(t *testing.T) {
		// Given
		ctx := t.Context()
		entry := logrus.NewEntry(logrus.New())
		entry = entry.WithField("CorrelationID", "123")
		ctxWithLogger := logger.ContextWithLogger(ctx, entry)

		// When
		retrievedLogger := logger.FromContext(ctxWithLogger)

		// Then
		require.Equal(t, entry, retrievedLogger)
	})

	t.Run("should return new logger with correlation ID when no logger in context", func(t *testing.T) {
		// Given
		ctx := t.Context()

		// When
		logger := logger.FromContext(ctx)

		// Then
		require.NotNil(t, logger)
		correlationID, ok := logger.Data["CorrelationID"]
		require.True(t, ok)
		require.NotEmpty(t, correlationID)
	})
}

func TestWithCorrelationID(t *testing.T) {
	t.Run("should create logger with unique correlation ID", func(t *testing.T) {
		// When
		logger1 := logger.WithCorrelationID()
		logger2 := logger.WithCorrelationID()

		// Then
		require.NotNil(t, logger1)
		require.NotNil(t, logger2)

		correlationID1, ok1 := logger1.Data["CorrelationID"]
		correlationID2, ok2 := logger2.Data["CorrelationID"]

		require.True(t, ok1)
		require.True(t, ok2)
		require.NotEmpty(t, correlationID1)
		require.NotEmpty(t, correlationID2)
		require.NotEqual(t, correlationID1, correlationID2)
	})
}
