package pair_with_error

import (
	"errors"
	"strconv"
	"sync"
)

type SourceM struct {
	Data  int // payload
	Error error
}

// Source is a stage that a pipeline starts from.
func Source(count int, errorAfter int) <-chan SourceM {
	out := make(chan SourceM)

	processor := func() {
		for n := 0; n < count; n++ {
			out <- SourceM{Data: n}

			if n == errorAfter { // Simulate some error
				out <- SourceM{Error: errors.New("Some error")}
				return
			}
		}
	}

	go func() {
		processor()
		close(out)
	}()
	return out
}

type MediatorM struct {
	Data  string // payload
	Error error
}

// Mediator is a stage somewhere in the middle of a pipeline. That usually
// grabs an input channel, processes it, and write the rusult to an output channel.
func Mediator(processors int, in <-chan SourceM) <-chan MediatorM {
	out := make(chan MediatorM)

	processor := func() {
		for in := range in {
			// Handle an error from the upstream. Common case here is just to
			// put the error down to the pipeline.
			if in.Error != nil {
				out <- MediatorM{Error: in.Error}
				continue // It's a mistake to stop here. The input channel must be fully exhausted!
			}

			out <- MediatorM{Data: strconv.Itoa(in.Data)}
		}
	}

	// Spawn the specified number of processors.
	// We need a waitgroup to wait for all the processors are done. After that
	// the output channel must be closed.
	wg := sync.WaitGroup{}
	for p := 0; p < processors; p++ {
		wg.Add(1)
		go func() { processor(); wg.Done() }()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Sink is a final stage in a pipeline. That usually collects the final result
// of the pipeline.
func Sink(in <-chan MediatorM) (result []string, err error) {
	for in := range in {
		if in.Error != nil {
			err = in.Error
			continue // It's a mistake to stop here. The input channel must be fully exhausted!
		}

		result = append(result, in.Data)
	}
	return
}

// Build and run a pipeline.
func Run(count int, errorAfter int) ([]string, error) {
	return Sink(
		Mediator(2,
			Source(count, errorAfter)))
}
