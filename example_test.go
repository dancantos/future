package future_test

import (
	"fmt"

	"github.com/dancantos/future"
)

func ExampleGo() {
	longRunningTask := func() string {
		return "Hello World"
	}
	value := future.Go(longRunningTask)
	fmt.Println(value.Get())

	// Output: Hello World
}

func ExampleGoErr() {
	longRunningError := func() (string, error) {
		return "", fmt.Errorf("Bad things")
	}
	value := future.GoErr(longRunningError)
	fmt.Println("value:" + value.Get().Value)
	fmt.Println("err: " + value.Get().Err.Error())

	// Output: value:
	// err: Bad things
}
