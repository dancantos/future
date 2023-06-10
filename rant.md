# Why futures?

Futures have long been avoided in go in favor of synchronization primitives,
goroutines, and channels. These work great for 95+% of use cases, but result in
horrible, untestable spaghetti code for the rest.

Consider a task where you need to perform 4 relatively long-running tasks, and
would like to parallelize them. C depends on A and D depends on A and B

A, B, C(A) and D(A,B)

How do we express this in code? We would like to say something like this...

```go
go func() { chA <- A() }()
go func() { chB <- B() }()
go func() { chC <- C(<-chA) }()
go func() { chD <- D(<-chA, <-chB) }()
```

This does not work. we are trying to read the result of A twice. Ok lets not start a routine until the values in needs are ready.

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

What if all of these can return errors.

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
chC, errChC := ChannelWrapper(func() {
    return C(valA)
})

if err := <-errChB; err != nil {
    // error
}
valB := <- chB
chD, errChC := ChannelWrapper(func() {
    return D(valA, valB)
})
```

This ChannelWrapper function essentially produces a future that can only be read once. However, as is, you MUST read both channels to avoid a goroutine leak, even if you don't care about one of the returns.

So lets just cut the crap and use proper futures.

```go
import "github.com/dancantos/future"

futureA := future.GoErr(A)
futureB := future.GoErr(B)

futureC := future.GoErr(func() {
    // C cares about whether A returned an error
    resultA := futureA.Get()
    if err := resultA.Err {
        // ...
    }
    return C(resultA.Value)
})

futureD := future.GoErr(func() {
    // D does not
    return D(futureA.Get().Value, futureB.Get().Value)
})
```

In this case, function D doesn't actually care if A or B returned errors.

