// Package bqueue is a "in memory" queue.
package bqueue

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
)

const (
	// DefaultLimit is the default job queue limit.
	DefaultLimit = 128

	// UnlimitedWorkers can passed to Workers when run in dynamic
	// mode to use an unlimited amount of workers.
	UnlimitedWorkers = -1
)

// Queue processes jobs.
type Queue struct {
	limiter chan struct{}
	workers int
	limit   int
	run     func() error
	jobs    chan Job
	log     Logger
	wg      sync.WaitGroup
}

// Option represents a queue option.
type Option func(q *Queue) error

// Workers sets the number of workers which will process jobs from the queue.
// The constant UnlimitedWorkers can be used along side dynamic workers, the default,
// which will enable unlimited workers.
// Default is runtime.NumCPU() workers.
func Workers(count int) Option {
	return func(q *Queue) error {
		if count < UnlimitedWorkers {
			return fmt.Errorf("workers must be at least %d, got %d", UnlimitedWorkers, count)
		}

		q.workers = count

		return nil
	}
}

// Limit sets the maximum number of jobs queued before an error is returned.
// Default is DefaultLimit maximum queued jobs..
func Limit(count int) Option {
	return func(q *Queue) error {
		if count < 0 {
			return fmt.Errorf("limit must be at least 0, got %d", count)
		}

		q.limit = count

		return nil
	}
}

// Log sets the logger for Queue.
// Default is log.Default().
func Log(l Logger) Option {
	return func(q *Queue) error {
		q.log = l
		return nil
	}
}

// Static configures the Queue to use a static number of pre-spawned
// workers instead of dynamic workers, which can be beneficial for
// inexpensive jobs.
// Default behaviour is to use dynamic workers.
func Static() Option {
	return func(q *Queue) error {
		q.run = q.static
		return nil
	}
}

// New creates a fully initialised Queue with the given options.
func New(options ...Option) (*Queue, error) {
	q := &Queue{
		workers: runtime.NumCPU(),
		limit:   DefaultLimit,
		log:     log.Default(),
	}
	q.run = q.dynamic

	for _, f := range options {
		if err := f(q); err != nil {
			return nil, err
		}
	}

	q.jobs = make(chan Job, q.limit)
	if err := q.run(); err != nil {
		return nil, err
	}

	return q, nil
}

// static spawns static workers.
func (q *Queue) static() error {
	if q.workers < 1 {
		return fmt.Errorf("workers must be at least 1, got %d", q.workers)
	}

	q.log.Printf("Spawning %d static workers...", q.workers)
	q.wg.Add(q.workers)

	for i := 0; i < q.workers; i++ {
		w := newWorker(q.jobs)
		go func() {
			defer q.wg.Done()
			w.run()
		}()
	}

	return nil
}

// dynamic processes jobs in goroutines.
// This optimises for the case where not all goroutines are needed all the time
// at the expense of having to start goroutines for each job.
func (q *Queue) dynamic() error {
	switch {
	case q.workers == UnlimitedWorkers:
		q.dynamicUnlimited()
	case q.workers < 1:
		return fmt.Errorf("workers must be at least 1, got %d", q.workers)
	default:
		q.dynamicLimited()
	}
	return nil
}

// dynamicUnlimited processes jobs with an unlimited number of goroutines.
// This optimises for the case where not all goroutines are needed all the time
// at the expense of having to start goroutines for each job.
func (q *Queue) dynamicUnlimited() {
	q.log.Printf("Using unlimited dynamic workers...", q.workers)
	q.wg.Add(1)

	go func() {
		defer q.wg.Done()

		for j := range q.jobs {
			q.wg.Add(1)
			go func(j Job) {
				defer q.wg.Done()
				j.Process()
			}(j)
		}
	}()
}

// dynamicLimited processes jobs in a limited number of goroutines.
// This optimises for the case where not all goroutines are needed all the time
// at the expense of having to start goroutines for each job.
func (q *Queue) dynamicLimited() {
	q.log.Printf("Using up to %d dynamic workers...", q.workers)
	q.wg.Add(1)
	q.limiter = make(chan struct{}, q.workers)

	go func() {
		defer q.wg.Done()

		for j := range q.jobs {
			q.limiter <- struct{}{}
			q.wg.Add(1)
			go func(j Job) {
				defer func() {
					<-q.limiter
					q.wg.Done()
				}()
				j.Process()
			}(j)
		}
	}()
}

// Queue queues a job in blocking mode.
func (q *Queue) Queue(job Job) {
	q.jobs <- job
}

// QueueNonBlocking queues a job in non blocking mode.
// If the maximum buffer as defined by Limit is already filled an error will be returned.
func (q *Queue) QueueNonBlocking(job Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return fmt.Errorf("too busy, already have %d queued jobs", cap(q.jobs))
	}
}

// Stop stops processing and returns once all jobs have been completed
// or the context indicates done.
// The queue should not be used after calling Stop, calling Queue or QueueNonBlocking
// after Stop will cause a panic.
func (q *Queue) Stop(ctx context.Context) error {
	close(q.jobs)
	done := make(chan struct{})
	go func() {
		defer close(done)
		q.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
