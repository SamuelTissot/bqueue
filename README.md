BQUEUE
====

##### Build Status

| branch   | status |
| -------- |:------:|
| master   | [![CircleCI](https://circleci.com/gh/SamuelTissot/bqueue/tree/master.svg?style=svg)](https://circleci.com/gh/SamuelTissot/bqueue/tree/master) |



**A buffered async queue**
###### Based on Marcio Castilho [article](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/) “Handling 1 Million Requests per Minute with Go”

Why
---
We needed a simple and quick queue system to handle requests rapidly without impacting performance.
Each of our jobs could take up to 20s to execute.

How it works
----
By using the awesomeness of channels.

>Channels are the pipes that all goroutines to share data. You can send values into channels from one goroutine and receive those values in another goroutine.

With channels we are able to share the data between goroutines enabling it to be processed concurrently, to make the best use of multiple CPU cores.


Usage
---

```go
	// Create the Queue with 2 static worker
	q, err := bqueue.New(bqueue.Static(), bqueue.Workers(2))
	if err != nil {
		// handle error
		panic(err)
	}

	// add jobs
    // any implementation of the bqueue.Job interface
	q.Queue(...)

	// will wait until all jobs are done to STOP or context is cancelled
	if err := q.Stop(context.TODO()); err != nil {
		// handle error
		panic(err)
	}
```

Example
---

- [Basic Example](./examples/basic)
