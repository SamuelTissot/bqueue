package bqueue

// Job is an interface for the bqueue.
// it must implement `func Process() err {}`
type Job interface {
	Process() error
}
