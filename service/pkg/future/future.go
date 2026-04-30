// Package future provides a small abstraction for starting work now and reading
// its single result later with timeout and cancellation support.
package future

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type (
	taskResult[R any] struct {
		val R
		err error
	}

	// Task represents one background operation started with NewTask.
	// It exists to decouple when work begins from when the caller needs the result.
	Task[R any] struct {
		mu       sync.Mutex
		consumed bool // true after Get successfully reads the result
		canceled bool // true after an explicit Cancel() call

		// resultCh stores the task outcome until a caller retrieves it with Get.
		resultCh chan taskResult[R]

		ctx    context.Context
		cancel context.CancelFunc
	}
)

var (
	ErrContextClosed  = errors.New("context closed before the task is completed")
	ErrTimeoutExpired = errors.New("timeout expired")
	ErrResultConsumed = errors.New("task result was already consumed")
	ErrCanceled       = errors.New("task is canceled")
)

// TaskError wraps an error returned by the task function itself.
// This lets callers distinguish task execution failures from waiting failures
// such as timeout, cancellation, or context closure.
type TaskError struct{ baseErr error }

func (e TaskError) Error() string { return fmt.Sprintf("task error: %v", e.baseErr) }
func (e TaskError) Unwrap() error { return e.baseErr }

// NewTask starts run in the background and returns a handle for waiting,
// cancellation, and simple state checks.
func NewTask[R any](ctx context.Context, run func(context.Context) (R, error)) *Task[R] {
	ctx, cancel := context.WithCancel(ctx)

	t := &Task[R]{
		resultCh: make(chan taskResult[R], 1),
		ctx:      ctx,
		cancel:   cancel,
	}

	go func() {
		defer cancel()

		// Bail out early if the context was already done before we started.
		if ctx.Err() != nil {
			return
		}

		val, err := run(ctx)
		res := taskResult[R]{val: val, err: err}

		t.mu.Lock()
		if t.canceled {
			t.mu.Unlock()
			return
		}
		t.resultCh <- res
		t.mu.Unlock()
	}()

	return t
}

// finishConsume transitions the task into the consumed state, tears down
// the worker context, and unwraps the result.
func (t *Task[R]) finishConsume(res taskResult[R]) (R, error) {
	var empty R

	t.mu.Lock()
	t.consumed = true
	t.mu.Unlock()

	t.cancel()

	if res.err != nil {
		return empty, TaskError{baseErr: res.err}
	}

	return res.val, nil
}

// stateError safely reads the task state and returns the corresponding error.
// If the task completed normally (neither canceled nor consumed), it returns fallback.
func (t *Task[R]) stateError(fallback error) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch {
	case t.canceled:
		return ErrCanceled
	case t.consumed:
		return ErrResultConsumed
	default:
		return fallback
	}
}

// Get waits up to timeout for the task result.
// Use it at the point where the caller actually needs the computed value.
// A timeout does not stop the task; the work keeps running until it
// finishes, the task is canceled via Cancel, or the parent context is done.
func (t *Task[R]) Get(timeout time.Duration) (R, error) {
	var empty R

	if err := t.stateError(nil); err != nil {
		return empty, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// Fast path: result is already buffered.
	select {
	case res := <-t.resultCh:
		return t.finishConsume(res)
	default:
	}

	select {
	case res := <-t.resultCh:
		return t.finishConsume(res)

	case <-timer.C:
		// Last chance: result may have landed between the timer firing and now.
		select {
		case res := <-t.resultCh:
			return t.finishConsume(res)
		default:
			return empty, ErrTimeoutExpired
		}

	case <-t.ctx.Done():
		// The worker defers cancel() after sending, so the result may already
		// be buffered. Drain before concluding the task did not complete.
		select {
		case res := <-t.resultCh:
			return t.finishConsume(res)
		default:
		}

		return empty, t.stateError(ErrContextClosed)
	}
}

// IsCompleted reports whether the task result has already been retrieved.
func (t *Task[R]) IsCompleted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.consumed
}

// IsCanceled reports whether the task was explicitly canceled.
func (t *Task[R]) IsCanceled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.canceled
}

// Cancel marks the task as canceled and closes its context for the running work.
func (t *Task[R]) Cancel() {
	t.mu.Lock()
	if t.canceled || t.consumed {
		t.mu.Unlock()
		return
	}
	t.canceled = true

	select {
	case <-t.resultCh:
	default:
	}
	t.mu.Unlock()

	t.cancel()
}
