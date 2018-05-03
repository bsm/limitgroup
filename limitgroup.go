package limitgroup

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Group is a limiting errgroup
type Group struct {
	limiter chan struct{}
	parent  *errgroup.Group
	ctx     context.Context
}

// New inits a new group with a given concurrency (default: 1)
func New(ctx context.Context, concurrency int) *Group {
	if concurrency < 1 {
		concurrency = 1
	}

	parent, ctx := errgroup.WithContext(ctx)
	return &Group{
		limiter: make(chan struct{}, concurrency),
		parent:  parent,
		ctx:     ctx,
	}
}

// Go runs a function as part of the group. The first function to
func (g *Group) Go(fn func() error) {
	select {
	case <-g.Done():
		return
	case g.limiter <- struct{}{}:
	}

	g.parent.Go(func() error {
		err := fn()

		select {
		case <-g.limiter:
		case <-g.Done():
		}
		return err
	})
}

// Wait waits for all the group functions to exit and returns the first error.
func (g *Group) Wait() error {
	return g.parent.Wait()
}

// Done is triggered when the group is done.
func (g *Group) Done() <-chan struct{} {
	return g.ctx.Done()
}
