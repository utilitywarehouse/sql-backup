package pool

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrZeroItems = errors.New("zero items provided")
)

type Handler func(ctx context.Context, db string) error

type Pooler interface {
	Start(context.Context, []string, Handler) error
}

type SizablePool struct {
	Size int
}

func (p SizablePool) Start(ctx context.Context, items []string, h Handler) error {
	if len(items) == 0 {
		return ErrZeroItems
	}

	var wg sync.WaitGroup
	wg.Add(len(items))

	jobs := make(chan string)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	finishCtx, finishCancel := context.WithCancel(ctx)
	defer finishCancel()

	if p.Size < 1 {
		p.Size = 1
	}
	for i := 0; i < p.Size; i++ {
		go func() {
			for job := range jobs {
				func(t string) {
					defer wg.Done()
					if err := h(finishCtx, job); err != nil {
						errCh <- err
					}
				}(job)
			}
		}()
	}

	go func() {
		for _, item := range items {
			jobs <- item
		}
		close(jobs)
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	case <-doneCh:
		return nil
	}
}
