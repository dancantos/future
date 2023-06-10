package future_test

import (
	"sync"
	"testing"

	"github.com/dancantos/future"
)

func TestFuture(t *testing.T) {
	t.Run("multiple_get_returns_value", func(t *testing.T) {
		testfn := func() bool {
			return true
		}
		boolFuture := future.Go(testfn)

		for i := 0; i < 10; i++ {
			assertTrue(t, boolFuture.Get(), "get did not return 'true'")
		}
	})

	t.Run("get_waits_for_value", func(t *testing.T) {
		t.Parallel()

		// This is a complicated concurrency test case. We want to make sure
		// Get() blocks until a value is ready, so we have a 'setter' routine
		// waiting for a signal on a channel that we trigger when the test is
		// ready
		//
		// To test the result of get, we use a second 'wait' routine that
		// blocks on Get() and sets a secondary bool value with the result.
		//
		// the test inspects this secondary value before and after sending the
		// signal to the setter. The change in this value indicates that Get()
		// has received the correct value from the setter.
		//
		// In the middle, we will also run a concurrency test, with 10 routines waiting for
		// a true value.

		// set the signal channel and setter routing
		signal := make(chan struct{})
		defer close(signal)
		testfn := func() bool {
			<-signal
			return true
		}
		boolFuture := future.Go(testfn)

		// set up the 'wait' routine
		// the waitgroup is so we can wait for this second routine to exit.
		var boolVal bool
		wg := sync.WaitGroup{}
		// set up the 10 routines
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				result := boolFuture.Get()
				assertTrue(t, result, "concurrent routine failed to Get() the correct set value")
				// only 1 routine should set the bool value
				if i == 0 {
					boolVal = result
				}
			}(i)
		}

		// Now we test
		assertFalse(t, boolVal, "Get() return incorrect value before correct value was set")
		signal <- struct{}{} // trigger the setter
		wg.Wait()            // wait for the second routine to set the secondary value
		assertTrue(t, boolVal, "Get() did not return the correct set value")
	})
}

func assertFalse(t *testing.T, boolValue bool, errmsg string) {
	if boolValue { // we expect false, true is a failed test
		t.Error(errmsg)
	}
}

func assertTrue(t *testing.T, boolValue bool, errmsg string) {
	if !boolValue {
		t.Error(errmsg)
	}
}
