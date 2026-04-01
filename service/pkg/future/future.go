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

	// Task represents one background operation started with `NewTask`.
	// It exists to decouple when work begins from when the caller needs the result.
	Task[R any] struct {
		mu       sync.Mutex
		consumed bool // true after Get successfully reads the result
		canceled bool // true after an explicit Cancel() call

		// resultCh stores the task outcome until a caller retrieves it with `Get`.
		resultCh chan taskResult[R]

		ctx    context.Context
		cancel context.CancelFunc
	}

	runFunc[R any] func(context.Context) (R, error)
)

var (
	ErrContextClosed  = errors.New("context closed before the task is completed")
	ErrTimeoutExpired = errors.New("timeout expired")
)

// TaskError wraps an error returned by the task function itself.
// This lets callers distinguish task execution failures from waiting failures
// such as timeout, cancellation, or context closure.
type TaskError struct{ baseErr error }

func (e TaskError) Error() string { return fmt.Sprintf("task error: %s", e.baseErr.Error()) }
func (e TaskError) Unwrap() error { return e.baseErr }

// NewTask starts `run` in the background and returns a handle for waiting,
// cancellation, and simple state checks.
func NewTask[R any](ctx context.Context, run runFunc[R]) *Task[R] {
	ctx, cancel := context.WithCancel(ctx)

	t := &Task[R]{
		resultCh: make(chan taskResult[R], 1),
		ctx:      ctx,
		cancel:   cancel,
	}

	go func() {
		// Bail out early if the context was already done before we started.
		if ctx.Err() != nil {
			return
		}
		val, err := run(ctx)
		// resultCh is buffered(1) and the producer is the sole sender, so this
		// send never blocks regardless of whether the result is ever consumed.
		t.resultCh <- taskResult[R]{val: val, err: err}
	}()

	return t
}

// Get waits up to `timeout` for the task result.
// Use it at the point where the caller actually needs the computed value.
func (t *Task[R]) Get(timeout time.Duration) (R, error) {
	var empty R

	t.mu.Lock()
	switch {
	case t.canceled:
		t.mu.Unlock()
		return empty, errors.New("task is canceled")
	case t.consumed:
		t.mu.Unlock()
		return empty, errors.New("task is completed")
	}
	t.mu.Unlock()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case res := <-t.resultCh:
		t.mu.Lock()
		t.consumed = true
		t.mu.Unlock()
		t.cancel() // release context resources now that the result is consumed
		if res.err != nil {
			return empty, TaskError{baseErr: res.err}
		}
		return res.val, nil

	case <-t.ctx.Done():
		t.mu.Lock()
		isCanceled := t.canceled
		t.mu.Unlock()
		if isCanceled {
			return empty, errors.New("task is canceled")
		}
		return empty, ErrContextClosed

	case <-timer.C:
		return empty, ErrTimeoutExpired
	}
}

// IsCompleted reports whether the task result has already been retrieved.
func (t *Task[R]) IsCompleted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.consumed
}

// IsCancelled reports whether the task was explicitly cancelled.
func (t *Task[R]) IsCancelled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.canceled
}

// Cancel marks the task as cancelled and closes its context for the running work.
func (t *Task[R]) Cancel() {
	t.mu.Lock()
	if t.canceled || t.consumed {
		t.mu.Unlock()
		return
	}
	t.canceled = true
	t.mu.Unlock()
	t.cancel()
}
