# go-aqueue

[![Build Status](https://travis-ci.org/dtoubelis/go-aqueue.svg?branch=master)](https://travis-ci.org/dtoubelis/go-aqueue)
[![codecov](https://codecov.io/gh/dtoubelis/go-aqueue/branch/master/graph/badge.svg)](https://codecov.io/gh/dtoubelis/go-aqueue)
[![GoReport](https://goreportcard.com/badge/github.com/dtoubelis/go-aqueue)](https://goreportcard.com/badge/github.com/dtoubelis/go-aqueue)
[![GoDoc](https://godoc.org/github.com/dtoubelis/go-aqueue?status.svg)](https://godoc.org/github.com/dtoubelis/go-aqueue)
![GitHub](https://img.shields.io/github/license/dtoubelis/go-aqueue)

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
better version.

At the core of MOP is an asynchronows queue and Go channel is a version of
it. The asyncronous queue used in MOP may resemble some known concurrency
patterns (pub/sub, pipeline, etc.) but it has different requirements:

- It is intended to communicae between exactly two parties (i.e. fair queuing
  is out of scope)
- It should not (arguably) provide any buffering. This requirement is based on
  personal experience and I'm yet to see a use case where buffering provides
  any advantages.

## Comparison with go channels

Here is a list of advantages and disadvantages of this package in comparison
with go channels:

**Pros:**

- AQueue can be gracefully closed from any thread and both reader and
  writer will exit with an error instead of panic.
- AQueue provides blocking as well as non-blocking methods for wrining to the
  queue and reading from it.
- Some AQueue methods return `cancel()` function that can be used for
  implementing timeouts (i.e. in conjunction with `context.WithDeadline()`
  or `context.WithTimeout()`) or gracefully closing the queue on
  `context.Done()`.

**Cons:**

- Type safety can only be assured at run time, mostly because of lack of
  generics in golang (as of Aug 2020).
- Performance of AQueue is 2-5 times slower than that of native channels.
  The difference in performance depends on hardware and there are benchmarks
  included with the package that will show the actual difference on your
  particular hardware.

## ToDo

- Provide more code examples
- Improve documentation

## About the Author

The author of this package has extensive experience implementing MOP concept
from highly efficient multimedia processing on low powered embedded devices
to parallel execution of thousands of cloud based applications performing
financial modeling.

## License

See [LICENSE](LICENSE).
