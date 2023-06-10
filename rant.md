# Why futures?

Futures have long been avoided in go in favor of synchronization primitives, goroutines, and channels.

- [no package required](https://appliedgo.net/futures/)
- [Use go channels as promises](https://levelup.gitconnected.com/use-go-channels-as-promises-and-async-await-ee62d93078ec)

These work great for 95+% of use cases, but result in horrifying spaghetti code for the rest.

## Problem statement

Consider a task where you need to perform 4 relatively long-running tasks, and would like to parallelize them. C depends on A and D depends on A and B. This is close to minimum level complexity that requires the use of futures.

A, B, C(A), D(A,B)

How do we express this in code? We would like to say something like this...

```go
go func() { chA <- A() }()
go func() { chB <- B() }()
go func() { chC <- C(<-chA) }()
go func() { chD <- D(<-chA, <-chB) }()
```

This does not work. we are trying to read the result of A twice. Ok lets not start a routine until the values it needs are ready.

```go
go func() { chA <- A() }()
go func() { chB <- B() }()
valA := <-chA
go func() { chC <- C(valA) }()
valB := <-chB
go func() { chD <- D(valA, valB) }()
```

But wait, we are missing a lot of code here

```go
chA := make(chan A)
chB := make(chan B)
chC := make(chan C)
chD := make(chan D)
go func() { chA <- A() }()
go func() { chB <- B() }()
valA := <-chA
go func() { chC <- C(valA) }()
valB := <-chB
go func() { chD <- D(valA, valB) }()
```

What if all of these can return errors?

```go
chA := make(chan A)
errChA := make(chan error)
chB := make(chan B)
errChB := make(chan error)
chC := make(chan C)
errChC := make(chan error)
chD := make(chan D)
errChD := make(chan error)
go func() {
    valA, errA := A()
    chA <- valA
    errChA <- errA
}()
go func() {
    valB, errB := A()
    chB <- valB
    errChB <- errB
}()
valA, errA := <-chA, errChA
if errA != nil {
    // what do?
}
go func() {
    ...
}()
```

We are already at the point where we might want to make some generic functions to handle a lot of this bloat. This one handles creation of the channes and running the goroutines.

```go
func ChannelWrapper[A any](fn() (A, error)) (chan A, chan error) {
    chA := make(chan A)
    errCh := make(chan error)
    go func() {
        val, err := fn()
        chA <- val
        errCh <- err
    }()
    return chA, errCh
}

chA, errChA := ChannelWrapper(A)
chB, errChB := ChannelWrapper(B)

if err := <-errChA; err != nil {
    // error
}
valA := <-chA
chC, errChC := ChannelWrapper(func() (CType, error) {
    return C(valA)
})

if err := <-errChB; err != nil {
    // error
}
valB := <- chB
chD, errChC := ChannelWrapper(func() (DType, error) {
    return D(valA, valB)
})
```

This ChannelWrapper function essentially produces a future that can only be read once. However, as is, you MUST read both channels to avoid a goroutine leak, even if you don't care about one of the returns.

## Cut the crap

So lets just cut the crap and use proper futures.

```go
import "github.com/dancantos/future"

futureA := future.GoErr(A)
futureB := future.GoErr(B)

futureC := future.GoErr(func() (CType, error) {
    // C cares about whether A returned an error
    resultA := futureA.Get()
    if err := resultA.Err {
        // ...
    }
    return C(resultA.Value)
})

futureD := future.GoErr(func() (DType, error) {
    // D does not
    return D(futureA.Get().Value, futureB.Get().Value)
})
```

In this case, function D doesn't actually care if A or B returned errors.

## Surely someone has done this before right?

A [search on pkg.go.dev](https://pkg.go.dev/search?q=future&m=) does not look promising, but they do exist.

- [Allan-Jacobs](https://pkg.go.dev/github.com/Allan-Jacobs/go-futures)
- [stephennancekivell](https://pkg.go.dev/github.com/stephennancekivell/go-future)
- [jbowes](https://pkg.go.dev/github.com/jbowes/future)
- [fanliao](https://pkg.go.dev/github.com/fanliao/go-promise)

But I didn't like any of them. [stephennancekivell](https://pkg.go.dev/github.com/stephennancekivell/go-future) is probably the closest to this library.

## What I don't want to see

Using futures is a slippery slope to effectively introducing the `async` keyword into a language. Any function that accepts futures as arguments must either return a future, or block until the input futures are ready. If these limitations are not communicated effectively to consumers, they risk abusing them and creating very hard bugs to debug.

With that said, here is my 1 rule for using futures...

> A future must never leave the function it is created in.

If your code base, a function can never accept a future as an argment, or return a future. For this reason, this library does NOT provide a method to create futures from values.
