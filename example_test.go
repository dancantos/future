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

func Example() {
	long1 := func() string {
		return "input2"
	}
	long2 := func(string) (int, error) {
		return 1, nil
	}

	future1 := future.Go(long1)
	future2 := future.GoErr(func() (int, error) {
		return long2(future1.Get())
	})

	if future2.Get().Err != nil {
		fmt.Println("uh oh")
	}

	fmt.Printf("value: %d\n", future2.Get().Value)

	// Output: value: 1
}
