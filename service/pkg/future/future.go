package future

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type (
	Task[R any] interface {
		Get(duration time.Duration) (R, error)
		IsCompleted() bool
		IsCancelled() bool
		Cancel()
	}

	task[R any] struct {
		completed bool
		canceled  bool
		closed    bool

		resultChan chan R
		errChan    chan error

		ctx    context.Context
		cancel func()
	}

	runFunc[R any] func(context.Context) (R, error)
)

func NewTask[R any](ctx context.Context, run runFunc[R]) Task[R] {
	ctx, cancel := context.WithCancel(ctx)

	task := &task[R]{
		completed:  false,
		canceled:   false,
		resultChan: make(chan R, 1),
		errChan:    make(chan error, 1),
		ctx:        ctx,
		cancel:     cancel,
	}

	task.run(ctx, run)

	return task
}

func (t *task[R]) Get(timeout time.Duration) (R, error) {
	var empty R

	switch {
	case t.canceled:
		return empty, errors.New("task is canceled")
	case t.closed:
		return empty, errors.New("task is completed")
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := t.wait(ctxWithTimeout)
	if err != nil && !errors.As(err, &TaskError{}) {
		return empty, err
	}

	t.close()

	return res, err
}

var (
	ErrContextClosed  = errors.New("context closed before the task is completed")
	ErrTimeoutExpired = errors.New("timeout expired")
)

type TaskError struct{ baseErr error }

func (e TaskError) Error() string { return fmt.Sprintf("task error: %s", e.baseErr.Error()) }
func (e TaskError) Unwrap() error { return e.baseErr }

func newTaskError(err error) error { return TaskError{baseErr: err} }

func (t *task[R]) wait(ctx context.Context) (R, error) {
	var empty R
	for {
		select {
		case <-t.ctx.Done():
			return empty, ErrContextClosed
		case <-ctx.Done():
			return empty, ErrTimeoutExpired
		case err := <-t.errChan:
			return empty, newTaskError(err)
		case res := <-t.resultChan:
			return res, nil
		}
	}
}

func (t *task[R]) run(ctx context.Context, run runFunc[R]) {
	go func(t *task[R]) {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := run(ctx)
			if err != nil {
				t.errChan <- err
			} else {
				t.resultChan <- res
			}
			t.complete()
		}
	}(t)
}

func (t *task[R]) IsCompleted() bool {
	return t.completed
}

func (t *task[R]) IsCancelled() bool {
	return t.canceled
}

func (t *task[R]) Cancel() {
	if t.canceled || t.completed {
		return
	}
	t.canceled = true

	t.close()
}

func (t *task[R]) complete() {
	if t.canceled || t.completed {
		return
	}
	t.completed = true
}

func (t *task[R]) close() {
	if t.closed {
		return
	}

	t.cancel()

	close(t.resultChan)
	close(t.errChan)

	t.closed = true
}
