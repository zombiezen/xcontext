// Copyright 2024 Ross Light
// SPDX-License-Identifier: BSD-3-Clause

package xcontext

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCloseWhenDone(t *testing.T) {
	t.Run("Background", func(t *testing.T) {
		ctx := context.Background()
		c1 := newFakeCloser(nil)

		c2 := CloseWhenDone(ctx, c1)
		if got, want := c1.count(), 0; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		if err := c2.Close(); err != nil {
			t.Error("Unexpected error from returned closer:", err)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}
	})

	t.Run("Error", func(t *testing.T) {
		ctx := context.Background()
		myError := errors.New("bork")
		c1 := newFakeCloser(myError)

		c2 := CloseWhenDone(ctx, c1)
		if got, want := c1.count(), 0; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		if err := c2.Close(); err != myError {
			t.Errorf("Close() = %v; want %v", err, myError)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}
	})

	t.Run("CloseTwice", func(t *testing.T) {
		ctx := context.Background()
		myError := errors.New("bork")
		c1 := newFakeCloser(myError)

		c2 := CloseWhenDone(ctx, c1)
		if got, want := c1.count(), 0; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		if err := c2.Close(); err != myError {
			t.Errorf("Close() = %v; want %v", err, myError)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		if err := c2.Close(); err != myError {
			t.Errorf("Close() = %v; want %v", err, myError)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}
	})

	t.Run("AlreadyDone", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		myError := errors.New("bork")
		c1 := newFakeCloser(myError)

		c2 := CloseWhenDone(ctx, c1)
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		if err := c2.Close(); err != myError {
			t.Errorf("Close() = %v; want %v", err, myError)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}
	})

	t.Run("DoneBeforeClose", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		myError := errors.New("bork")
		c1 := newFakeCloser(myError)

		c2 := CloseWhenDone(ctx, c1)
		if got, want := c1.count(), 0; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		cancel()
		if got, want := c1.waitForClose(), 1; got != want {
			t.Errorf("Close() called %d times after cancel; want %d", got, want)
		}

		if err := c2.Close(); err != myError {
			t.Errorf("Close() = %v; want %v", err, myError)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}
	})

	t.Run("DoneAfterClose", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		myError := errors.New("bork")
		c1 := newFakeCloser(myError)

		c2 := CloseWhenDone(ctx, c1)
		if got, want := c1.count(), 0; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		if err := c2.Close(); err != myError {
			t.Errorf("Close() = %v; want %v", err, myError)
		}
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times; want %d", got, want)
		}

		cancel()
		// Not deterministic that it will catch the defect,
		// but we're checking for the absence of something occuring.
		time.Sleep(1 * time.Millisecond)
		if got, want := c1.count(), 1; got != want {
			t.Errorf("Close() called %d times after cancel; want %d", got, want)
		}
	})
}

type fakeCloser struct {
	err error

	mu         sync.Mutex
	cond       sync.Cond
	closeCount int
}

func newFakeCloser(err error) *fakeCloser {
	c := &fakeCloser{err: err}
	c.cond.L = &c.mu
	return c
}

func (c *fakeCloser) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closeCount
}

func (c *fakeCloser) waitForClose() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.closeCount == 0 {
		c.cond.Wait()
	}
	return c.closeCount
}

func (c *fakeCloser) Close() error {
	c.mu.Lock()
	c.closeCount++
	c.cond.Broadcast()
	c.mu.Unlock()
	return c.err
}
