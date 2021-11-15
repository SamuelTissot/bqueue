package bqueue

// Job is implemented by types which can be processed by Queue.
type Job interface {
	Process()
}

// Logger is implemented by types which an be used by Queue as a log destination.
type Logger interface {
	Printf(format string, v ...interface{})
}
