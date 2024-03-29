package future

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFutureTaskCorrect(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run := func(ctx context.Context) (string, error) { return "done", nil }

	task := NewTask[string](ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	res, err := task.Get(time.Second)
	require.NoError(t, err)
	require.Equal(t, "done", res)
	require.False(t, task.IsCancelled())
	require.True(t, task.IsCompleted())

	res, err = task.Get(time.Second)
	require.ErrorContains(t, err, "task is completed")
	require.Equal(t, "", res)
}

func TestFutureTaskContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	run := func(ctx context.Context) (string, error) {
		for {
			select {
			case <-ctx.Done():
				return "", nil
			default:
				time.Sleep(time.Second)
				return "done", nil
			}
		}
	}

	task := NewTask[string](ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	cancel()

	res, err := task.Get(time.Second)
	require.ErrorContains(t, err, "context closed before the task is completed")
	require.Equal(t, "", res)
	require.False(t, task.IsCompleted())
}

func TestFutureTaskTimeoutExpired(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run := func(ctx context.Context) (string, error) {
		for {
			select {
			case <-ctx.Done():
				return "", nil
			default:
				time.Sleep(2 * time.Second)
				return "done", nil
			}
		}
	}

	task := NewTask[string](ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	res, err := task.Get(time.Second)
	require.ErrorContains(t, err, "timeout expired")
	require.Equal(t, "", res)
	require.False(t, task.IsCompleted())

	time.Sleep(2 * time.Second)

	res, err = task.Get(time.Second)
	require.NoError(t, err)
	require.Equal(t, "done", res)
	require.False(t, task.IsCancelled())
	require.True(t, task.IsCompleted())
}

func TestFutureTaskRunWithTaskError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run := func(ctx context.Context) (string, error) { return "", assert.AnError }

	task := NewTask[string](ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	res, err := task.Get(time.Second)
	require.ErrorAs(t, err, &TaskError{})
	require.ErrorIs(t, err, assert.AnError)
	require.Equal(t, "", res)
	require.False(t, task.IsCancelled())
	require.True(t, task.IsCompleted())
}

func TestFutureTaskRunTaskCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run := func(ctx context.Context) (string, error) {
		time.Sleep(time.Second)
		return "done", nil
	}

	task := NewTask[string](ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	task.Cancel()
	require.True(t, task.IsCancelled())

	res, err := task.Get(time.Second)
	require.ErrorContains(t, err, "task is canceled")
	require.Equal(t, "", res)
	require.True(t, task.IsCancelled())
	require.False(t, task.IsCompleted())
}
