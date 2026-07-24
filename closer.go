// Copyright 2024-2026 Roxy Light
// SPDX-License-Identifier: BSD-3-Clause

package xcontext

import (
	"context"
	"io"
)

type closer struct {
	closer        io.Closer
	stopAfterFunc func() bool

	closed chan struct{}
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
	cc := &closer{
		closer: c,
		closed: make(chan struct{}),
	}
	cc.stopAfterFunc = context.AfterFunc(ctx, cc.close)
	return cc
}

func (c *closer) Close() error {
	if c.stopAfterFunc() {
		c.close()
	} else {
		<-c.closed
	}
	return c.err
}

func (c *closer) close() {
	c.err = c.closer.Close()
	close(c.closed)
}
