package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type job struct{ x float32 }
type jobqueue chan job
type workerpool chan jobqueue

type donequeue chan bool
type quitqueue chan bool

var numberOfWorkers int
var maxQueue int
var theJobQueue jobqueue

var theDoneQueue donequeue

type dispatcher struct {
	s  string
	wp workerpool
	qq quitqueue
}

type worker struct {
	s  string
	wp workerpool
	jq jobqueue
	qq quitqueue
}

func initialize() {
	numberOfWorkers = 2
	maxQueue = 3
	theJobQueue = make(jobqueue, maxQueue)
	d := newDispatcher()
	d.run()

	fmt.Fprintf(os.Stdout, "theJobQueue was initialized with %d length.\n", maxQueue)
	theDoneQueue = make(donequeue)
	fmt.Fprintf(os.Stdout, "theDoneQueue is %v\n", theDoneQueue)
	fmt.Fprintf(os.Stdout, "---\n")
	time.Sleep(time.Second)
}

func newDispatcher() *dispatcher {
	wpool := make(workerpool, numberOfWorkers)
	fmt.Printf("The pool in the Dispatcher will have %d workers.\n", numberOfWorkers)

	fmt.Printf("The Dispatcher is initialized ... \n")
	return &dispatcher{
		s:  "dispatcher:0",
		wp: wpool,
		qq: make(chan bool, numberOfWorkers),
	}
}

func (d *dispatcher) run() {
	fmt.Fprintf(os.Stdout, "[%s] Running...\n", d.s)
	for i := 0; i < numberOfWorkers; i++ {
		id := fmt.Sprintf("workers::%s", strconv.Itoa(i))
		w := d.newWorker(id)
		w.start()
	}
	go d.dispatch()
}

func (d *dispatcher) dispatch() {
	fmt.Fprintf(os.Stdout, "[%s] Start dispatching messages for the workers, popping from theJobQueue %+v\n", d.s, theJobQueue)
	for {
		if j, more := <-theJobQueue; more {
			fmt.Fprintf(os.Stdout, "[%s] Popping... %d/%d jobs from queue\n", d.s, len(theJobQueue), maxQueue)

			// a job request has been received... but is blocked on the WorkerPool
			// channel thus is waiting for the first Worker free.
			go func(j job) {
				fmt.Fprintf(os.Stdout, "[%s] Job %+v received... taking a worker from %+v (deregistering)\n", d.s, j, d.wp)

				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jq := <-d.wp

				fmt.Fprintf(os.Stdout, "[%s] The job popped goes to the owner of the queue %+v. The pool is now %d long.\n", d.s, jq, len(d.wp))

				// dispatch the job to the worker job channel
				jq <- j
			}(j)

			time.Sleep(time.Second)
		} else {
			fmt.Fprintf(os.Stdout, "[%s] Received all jobs; notify stop and done...\n", d.s)
			for i := 0; i < numberOfWorkers; i++ {
				d.qq <- true
			}
			theDoneQueue <- true
			return
		}
	}
}

func (d *dispatcher) newWorker(id string) *worker {
	wc := make(jobqueue)
	fmt.Fprintf(os.Stdout, "[%s] Creating Worker from pool %+v with channel %+v\n", d.s, d.wp, wc)
	return &worker{
		s:  id,
		wp: d.wp,
		jq: wc,
		qq: d.qq,
	}
}

func (w *worker) start() {
	go func() {
		for {
			fmt.Printf("[%s] Registering: puts his jq %+v in the pool %+v...\n", w.s, w.jq, w.wp)
			w.wp <- w.jq

			select {
			case job := <-w.jq:
				fmt.Fprintf(os.Stdout, "[%s] %.1f popped from %+v (%d)...\n", w.s, job.x, w.jq, len(w.jq))
				time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			case <-w.qq:
				fmt.Fprintf(os.Stdout, "[%s] Received a stop signal... quitting	\n", w.s)
				return
			}
		}
	}()
}

func (w *worker) stop() {
	go func() {
		w.qq <- true
	}()
}

func main() {
	initialize()
	count := 10
	for i := 0; i < count; i++ {

		j := job{x: float32(i)}
		fmt.Printf("[---] Waiting for queueing a new job into theJobQueue\n")
		theJobQueue <- j
		fmt.Printf("[---] Word request queued... now %d/%d jobs in theJobQueue %+v\n", len(theJobQueue), maxQueue, theJobQueue)
	}

	close(theJobQueue)
	fmt.Fprintf(os.Stdout, "[---] Done queuing %d jobs... channel closed!\n", count)
	<-theDoneQueue
	fmt.Fprintf(os.Stdout, "\nBye\n")
}
