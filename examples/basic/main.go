package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/SamuelTissot/bqueue"
)

// AJob is a struct that implement the bqueue.Job interface
type AJob struct {
	id int
}

// Process implements bqueue.Job
func (j *AJob) Process() {
	// simulate work duration to show that the jobs are not sequentials
	duration := workDuration()

	time.Sleep(duration)

	fmt.Printf(
		"processed job: %d, job took: %d milliseconds\n",
		j.id,
		duration.Milliseconds(),
	)
}

func main() {

	// Create the Queue with 2 static worker
	q, err := bqueue.New(bqueue.Static(), bqueue.Workers(2))
	if err != nil {
		// handle error
		panic(err)
	}

	// blocking until all jobs are added
	addJobsAsync(q)

	// will wait until all jobs are done to STOP
	if err := q.Stop(context.TODO()); err != nil {
		// handle error
		panic(err)
	}
}

// addJobsAsync simulates job comming in randomly
// it will add 50 jobs
func addJobsAsync(q *bqueue.Queue) {
	ticker := time.NewTicker(10 * time.Millisecond)
	aID := atomic.Uint32{}

	for range ticker.C {
		id := int(aID.Add(1))

		if id > 50 {
			return
		}

		q.Queue(&AJob{
			id: id,
		})
	}
}

// workDuration is a helpeer function to simulate work time
func workDuration() time.Duration {
	mult := rand.Intn(1000-100) + 100
	return time.Millisecond * time.Duration(mult)
}
