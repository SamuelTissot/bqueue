package bqueue

// worker processes jobs.
type worker struct {
	jobs <-chan Job
}

// newWorker creates a new worker which processes from jobs until it's closed.
func newWorker(jobs <-chan Job) *worker {
	return &worker{jobs: jobs}
}

// run processes jobs until the jobs channel is closed..
func (w *worker) run() {
	for j := range w.jobs {
		j.Process()
	}
}
