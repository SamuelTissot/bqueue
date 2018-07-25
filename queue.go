// Package bqueue is a "in memory" queue.
//
// it Collects jobs via the `CollectJob` func who takes a
// interface `Job`
package bqueue

// Queue that process jobs reveived
type Queue struct {
	maxJobs     int
	JobRequests chan chan Job
	JobReceived chan Job
}

// New Queue object
// @param maxJobs int is the maximum job at a single time that can be handled
func New(maxJobs int) *Queue {
	JobRequests := make(chan chan Job, maxJobs)
	return &Queue{
		maxJobs:     maxJobs,
		JobRequests: JobRequests,
		JobReceived: make(chan Job, 128),
	}
}

// Start the queue
func (q *Queue) Start() {
	for i := 0; i < q.maxJobs; i++ {
		id := i + 1
		worker := newWorker(id, q.JobRequests)
		worker.start()
	}

	go q.dispatch()
}

// CollectJob Adds a job to the Queue
func (q *Queue) CollectJob(job Job) {
	q.JobReceived <- job
}

func (q *Queue) dispatch() {
	for {
		select {
		case job := <-q.JobReceived:
			go func() {
				jobRequest := <-q.JobRequests
				jobRequest <- job
			}()
		}
	}
}
