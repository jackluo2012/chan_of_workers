# Chan of Workers

A scaffolding for a simple job queue systems with a dispatcher and some workers
... written easily in Go - #channels #golang #queues

This is a very simple and (IMVHO) quite idiomatic example of how Go is great
in handling these asynchronous queue systems.

## How it works

The `main` initializes a `dispatcher` and start pushing jobs on a queue.
In the meantime the `dispatcher` has started some `workers`, takes the
jobs enqueued by the `main` and, using a simple pool (a channel of
channels), choose the first available `worker` to take care of the task

## Usage

`$ go run chan_of_workers.go`

## Contributing
Please, any advice or suggestion is truly appreciated!
