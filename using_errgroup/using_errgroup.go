package using_errgroup

import (
	"context"
	"errors"
	"strconv"

	"golang.org/x/sync/errgroup"
)

type Context struct {
	context.Context
	*errgroup.Group
}

func WithContext(ctx context.Context) *Context {
	gr, ctx := errgroup.WithContext(ctx)
	return &Context{ctx, gr}
}

// Run the specified number of processors in parallel,
// wait until all of them are done.
func Parallel(count int, processor func() error) error {
	var gr errgroup.Group
	for p := 0; p < count; p++ {
		gr.Go(processor)
	}
	return gr.Wait()
}

// Source is a stage that a pipeline starts from.
func Source(x *Context, count int) <-chan int {
	out := make(chan int)

	processor := func() error {
		for n := 1; n <= count; n++ {
			select {
			case <-x.Done():
				return x.Err()
			// You must alway use non-blocking send because a downstream might
			// never make a read from the channel and the goroutine will leak!
			case out <- n:
			}
		}
		return nil
	}

	// Example of a single-processing stage (single processor).
	x.Go(func() error {
		defer close(out)
		return processor()
	})
	return out
}

// Mediator is a stage somewhere in the middle of a pipeline. That usually
// grabs an input channel, processes it, and write the rusult to an output channel.
func Mediator(x *Context, processors int, errorAfter int, in <-chan int) <-chan string {
	out := make(chan string)

	processor := func() error {
		for in := range in {
			select {
			case <-x.Done():
				return x.Err()
			// You must alway use non-blocking send because a downstream might
			// never make a read from the channel and the goroutine will leak!
			case out <- strconv.Itoa(in):
			}

			if in >= errorAfter { // Simulate a error
				return errors.New("Some error")
			}
		}
		return nil
	}

	// Example of a mutli-processing stage (many proccessors).
	// We need to wait for all the processors are done. After that
	// the output channel must be closed.
	// CRITICAL: the processors must be run in a separate errgroup!
	x.Go(func() error {
		defer close(out)
		return Parallel(processors, processor)
	})
	return out
}

// Sink is a final stage in a pipeline. It usually collects the final result
// of the pipeline. Commonly it does make no sense to run it asyncly.
func Sink(x *Context, in <-chan string) (result []string, err error) {
	for in := range in {
		result = append(result, in)
	}
	return result, x.Wait()
}

// Build and run a pipeline.
func Run(ctx context.Context, count int, errorAfter int) ([]string, error) {
	x := WithContext(ctx)

	return Sink(x,
		Mediator(x, 2, errorAfter,
			Source(x, count)))

	// Or more explicitly:
	// numbers := Source(x, count)
	// strings := Mediator(x, 2, errorAfter, numbers)
	// return Sink(x, strings)
}
