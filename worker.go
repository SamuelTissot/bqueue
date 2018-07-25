package bqueue

import (
	"fmt"
)

type worker struct {
	ID          int
	JobHunting  chan Job
	jobRequests chan chan Job
	QuitChan    chan bool
}

func newWorker(id int, jobRequests chan chan Job) worker {
	return worker{
		ID:          id,
		JobHunting:  make(chan Job),
		jobRequests: jobRequests,
		QuitChan:    make(chan bool),
	}
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *worker) start() {
	go func() {
		for {
			// hunt for a job (advertise that we are ready to work
			w.jobRequests <- w.JobHunting
			select {
			case j := <-w.JobHunting:
				w.do(j)
			case <-w.QuitChan:
				fmt.Printf("worker%d: Stopping\n", w.ID)
				return
			}
		}
	}()
}

// stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w *worker) stop() {
	go func() {
		w.QuitChan <- true
	}()
}

func (w *worker) do(j Job) {
	err := j.Process()
	// todo handle errors
	if err != nil {
		fmt.Println(err)
	}
}
