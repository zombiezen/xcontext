// Copyright 2024 Ross Light
// SPDX-License-Identifier: BSD-3-Clause

package xcontext

import (
	"context"
	"io"
	"sync"
)

type closer struct {
	closed chan struct{}

	once   sync.Once
	closer io.Closer
	err    error
}

// CloseWhenDone calls c.Close() when the Context is Done
// or the returned [io.Closer] is called,
// whichever comes first.
// It guarantees that c.Close() will be called at most once.
// Subsequent calls to the the returned [io.Closer]'s Close method
// will return the error returned by c.Close().
//
// Closing the returned [io.Closer] releases resources associated with it,
// so code should close the returned [io.Closer] as soon as c is no longer being used.
func CloseWhenDone(ctx context.Context, c io.Closer) io.Closer {
	done := ctx.Done()
	if done == nil {
		// If the Context will never be Done, skip the goroutine.
		return &closer{closer: c}
	}
	select {
	case <-done:
		// If the Context is already done, close c and return the original context.
		err := c.Close()
		return nopCloser{err}
	default:
		cc := &closer{
			closer: c,
			closed: make(chan struct{}),
		}
		go func() {
			select {
			case <-ctx.Done():
				cc.Close()
			case <-cc.closed:
			}
		}()
		return cc
	}
}

func (c *closer) Close() error {
	c.once.Do(func() {
		if c.closed != nil {
			close(c.closed)
		}
		c.err = c.closer.Close()
	})
	return c.err
}

type nopCloser struct {
	err error
}

func (c nopCloser) Close() error {
	return c.err
}
