package future

import (
	"sync"
)

// Go starts the given function in a goroutine and returns a future to hold the
// results.
//
// The caller MUST make sure the routine exits eventually. If not, consumers of
// the future will block indefinitely.
func Go[T any](fn func() T) Future[T] {
	f := new[T]()
	go func() {
		value := fn()
		f.setValue(value)
	}()
	return f
}

// GoErr starts the given function in a goroutine and returns a future to hold
// the results.
//
// The caller MUST make sure the routine exits eventually. If not, consumers of
// the future will block indefinitely.
func GoErr[T any](fn func() (T, error)) Future[ValueOrError[T]] {
	f := new[ValueOrError[T]]()
	go func() {
		value, err := fn()
		f.setValue(ValueOrError[T]{value, err})
	}()
	return f
}

// Future represents the result of a computation that has not completed yet.
type Future[T any] interface {
	// Get blocks until the value is ready.
	//
	// NOTE: future does NOT guarantee immutability of the return value.
	Get() T
}

// future implements the Future[T] interface.
type future[T any] struct {
	cond  *sync.Cond
	value T
	set   bool
}

// new builds a new future
//
// Returned future can have its value set exactly once.
// Technically at this point, this is a promise to this package, but a future
// to consumers.
func new[T any]() *future[T] {
	return &future[T]{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

// used to set the value of a future by a goroutine.
//
// Set should only be called once. Subsequent calls are ignored.
func (f *future[T]) setValue(value T) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	if f.set {
		return
	}
	f.set = true
	f.value = value
	f.cond.Broadcast()
}

// Get implements Future[T].Get().
func (f *future[T]) Get() T {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	for !f.set {
		f.cond.Wait()
	}
	return f.value
}

// ValueOrError represents the result of a routine that returns a value OR an
// error.
type ValueOrError[T any] struct {
	Value T
	Err   error
}
