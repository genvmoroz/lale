package future_test

import (
	"context"
	"testing"
	"time"

	"github.com/genvmoroz/lale/service/pkg/future"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestFutureTaskCorrect(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	run := func(_ context.Context) (string, error) { return "done", nil }

	task := future.NewTask(ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())

	res, err := task.Get(time.Second)
	require.NoError(t, err)
	require.Equal(t, "done", res)
	require.False(t, task.IsCancelled())
	require.True(t, task.IsCompleted())

	res, err = task.Get(time.Second)
	require.ErrorContains(t, err, "task is completed")
	require.Empty(t, res)
}

func TestFutureTaskContextCanceled(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	run := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", nil
		default:
			time.Sleep(time.Second)
			return "done", nil
		}
	}

	task := future.NewTask(ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	res, err := task.Get(time.Second)
	require.ErrorContains(t, err, "context closed before the task is completed")
	require.Empty(t, res)
	require.False(t, task.IsCompleted())
}

func TestFutureTaskTimeoutExpired(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	run := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", nil
		default:
			time.Sleep(2 * time.Second)
			return "done", nil
		}
	}

	task := future.NewTask(ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	res, err := task.Get(time.Second)
	require.ErrorContains(t, err, "timeout expired")
	require.Empty(t, res)
	require.False(t, task.IsCompleted())

	select {
	case <-ctx.Done():
		t.Fatalf("context is closed, context error: %s", ctx.Err())
	case <-time.After(2 * time.Second):
	}

	res, err = task.Get(time.Second)
	require.NoError(t, err)
	require.Equal(t, "done", res)
	require.False(t, task.IsCancelled())
	require.True(t, task.IsCompleted())
}

func TestFutureTaskRunWithTaskError(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	run := func(_ context.Context) (string, error) { return "", assert.AnError }

	task := future.NewTask(ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	res, err := task.Get(time.Second)
	require.ErrorAs(t, err, &future.TaskError{})
	require.ErrorIs(t, err, assert.AnError)
	require.Empty(t, res)
	require.False(t, task.IsCancelled())
	require.True(t, task.IsCompleted())
}

func TestFutureTaskRunTaskCanceled(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	run := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
			return "done", nil
		}
	}

	task := future.NewTask(ctx, run)
	require.NotNil(t, task)

	require.False(t, task.IsCancelled())
	require.False(t, task.IsCompleted())

	task.Cancel()
	require.True(t, task.IsCancelled())

	res, err := task.Get(time.Second)
	require.ErrorContains(t, err, "task is canceled")
	require.Empty(t, res)
	require.True(t, task.IsCancelled())
	require.False(t, task.IsCompleted())
}

func TestFutureTaskConcurrency(t *testing.T) {
	defer goleak.VerifyNone(t)

	type testCase struct {
		name string
		run  func(ctx context.Context) (string, error)
		test func(t *testing.T, task *future.Task[string], ctx context.Context)
	}

	testCases := []testCase{
		{
			name: "cancel races with completion may panic on send to closed channel",
			run: func(ctx context.Context) (string, error) {
				// Give Cancel enough time to run close() before we send.
				time.Sleep(50 * time.Millisecond)
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				default:
					return "done", nil
				}
			},
			test: func(t *testing.T, task *future.Task[string], _ context.Context) {
				// If implementation closes channels inside Cancel while the producer
				// goroutine is still running, this sequence can trigger a
				// "send on closed channel" panic in the producer.
				task.Cancel()

				// Wait longer than the run func sleeps so that any panic in the
				// producer goroutine is likely to manifest during this test.
				time.Sleep(200 * time.Millisecond)
			},
		},
		{
			name: "concurrent get and cancel exercise data races",
			run: func(ctx context.Context) (string, error) {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(200 * time.Millisecond):
					return "done", nil
				}
			},
			test: func(t *testing.T, task *future.Task[string], _ context.Context) {
				const workers = 10
				errCh := make(chan error, workers)

				for range workers {
					go func() {
						_, _ = task.Get(500 * time.Millisecond)
						errCh <- nil
					}()
				}

				// Race Cancel against concurrent Get calls.
				go task.Cancel()

				for range workers {
					require.NoError(t, <-errCh)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			task := future.NewTask(ctx, tc.run)
			require.NotNil(t, task)

			tc.test(t, task, ctx)
		})
	}
}
