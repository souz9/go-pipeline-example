package using_errgroup

import (
	"context"
	"errors"
	"strconv"

	"golang.org/x/sync/errgroup"
)

// Source is a stage that a pipeline starts from.
func Source(
	ctx context.Context, gr *errgroup.Group,
	count int,
) <-chan int {
	out := make(chan int)

	// Example of a single-processing stage (single processor).
	gr.Go(func() error { // processor
		defer close(out)

		for n := 1; n <= count; n++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			// You must alway use non-blocking send because a downstream might
			// never make a read from the channel and the goroutine will leak!
			case out <- n:
			}
		}
		return nil
	})

	return out
}

// Mediator is a stage somewhere in the middle of a pipeline. That usually
// grabs an input channel, processes it, and write the rusult to an output channel.
func Mediator(
	ctx context.Context, gr *errgroup.Group,
	processors int, errorAfter int,
	in <-chan int,
) <-chan string {
	out := make(chan string)

	// Example of a mutli-processing stage (many proccessors).
	processor := func() error {
		for in := range in {
			select {
			case <-ctx.Done():
				return ctx.Err()
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

	// We need to wait for all the processors are done. After that
	// the output channel must be closed.
	// CRITICAL: the processors must be run in a separate errgroup!
	gr.Go(func() error {
		defer close(out)
		return spawnProcessors(processors, processor)
	})
	return out
}

// Spawn the specified number of processors in an rrgroup, wait until
// all of them are done.
func spawnProcessors(count int, processor func() error) error {
	var gr errgroup.Group
	for p := 0; p < count; p++ {
		gr.Go(processor)
	}
	return gr.Wait()
}

// Sink is a final stage in a pipeline. It usually collects the final result
// of the pipeline. Commonly it does make no sense to run it asyncly.
func Sink(ctx context.Context, gr *errgroup.Group, in <-chan string) (result []string, err error) {
	for in := range in {
		result = append(result, in)
	}
	return result, gr.Wait()
}

// Build and run a pipeline.
func Run(ctx context.Context, count int, errorAfter int) ([]string, error) {
	gr, ctx := errgroup.WithContext(ctx)

	return Sink(ctx, gr,
		Mediator(ctx, gr, 2, errorAfter,
			Source(ctx, gr, count)))

	// Or more explicitly:
	// numbers := Source(ctx, gr, count)
	// strings := Mediator(ctx, gr, 2, errorAfter, numbers)
	// return Sink(ctx, gr, strings)
}
