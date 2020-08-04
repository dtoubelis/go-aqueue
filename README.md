# go-aqueue

Go Asynchronous Queue is a better alternative to go channels. This package
requires `go1.8` or newer.

## Background

The Message Oriented Programing (or MOP) is not a novelty and is a concept
known for decades. This concept was implemented in many different languages
either natively or through various libraries and it is indispensable tool for
developing higly scalable and extremelly reliable applications and services.

It is nice to see an attempt from Go team to embed it into the language as
native feature. However, there is a number of shortcomings in the current
implementation of go channels and this package is an attemt to provide a
better implementation.

At the core of MOP is an asynchronows queue and Go channel is a version of
it. The asyncronous queue used in MOP may resemble some known concurrency
patterns (pub/sub, pipeline, etc.) but it has different requirements:

- It is intended to communicae between exactly two parties (i.e. fair queuing
  is out of scope)
- It should not provide any buffering. This requirement may be controversial
  but it is based on personal experience and I'm yet to see a use case where
  buffering provides any advantages.

## Comparison with go channels

Here is a list of advantages and disadvantages of this package in comparison
with go channels:

**Pros:**

- AQueue can be gracefully closed from any thread and both reader and
  writer will exit with an error message instead of panic.
- AQueue provides non-blocking methods for wrining to thequeue and reading
  from it.
- Some AQueue methods return cancel() function that can be used for
  implementing timeouts (i.e. in conjunction with `context.WithDeadline()`
  or `context.WithTimeout()`) or gracefully closing the queue on
  `context.Done()`

**Cons:**

- Type safety can only be assured at run time, mostly because of lack of
  generics in golang.
- Performance of AQueue is 2-5 times slower than that of native channels.
  The difference in performance depends on hardware and there are benchmarks
  included with the package that will show the actual difference on your
  particular hardware.

## ToDo

- Provide more code examples
- Improve documentation
- Integrate with one or more CI platforms (TravisCI, CircleCI) for testing
  on different platforms and producing badges.

## About the Author

The author of this package has extensive experience implementing MOP concept
from highly efficient multimedia processing on low powered embedded devices
to parallel execution of thouthands of cloud based applications performing
financial modeling.

## License

See [LICENSE](LICENSE).
