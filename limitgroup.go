// Package limitgroup provides a control structure for a group of goroutines with configurable concurrency (max number of goroutines to run in parallel).
package limitgroup

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Group is a concurrency-limiting errgroup.
type Group struct {
	limiter chan struct{}
	parent  *errgroup.Group
	ctx     context.Context
}

// New inits a new group with a given concurrency (default: 1).
func New(ctx context.Context, concurrency int) (*Group, context.Context) {
	if concurrency < 1 {
		concurrency = 1
	}

	parent, ctx := errgroup.WithContext(ctx)
	return &Group{
		limiter: make(chan struct{}, concurrency),
		parent:  parent,
		ctx:     ctx,
	}, ctx
}

// Go runs a function as part of the group.
// If group's context is cancelled, this is a no-op.
func (g *Group) Go(fn func() error) {
	select {
	case <-g.ctx.Done():
		return
	case g.limiter <- struct{}{}:
	}

	g.parent.Go(func() error {
		err := fn()

		select {
		case <-g.limiter:
		case <-g.ctx.Done():
		}
		return err
	})
}

// Wait waits for all the group functions to exit and returns the first error.
func (g *Group) Wait() error {
	return g.parent.Wait()
}
