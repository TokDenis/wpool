package wpool

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	Free    = 0
	InWork  = 1
	Errored = 2
)

type worker struct {
	id                     int64
	workerContractsJobChan chan Job
	status                 int64

	jobNameInWork string
	started       time.Time
	elapsed       time.Duration
	m             sync.Mutex
}

type workerMetric struct {
	jobName string
	state   int64
	started time.Time
	elapsed time.Duration
}

func (w *worker) consume(j Job) {
	w.workerContractsJobChan <- j
}

func (w *worker) work() {
	for {
		atomic.SwapInt64(&w.status, Free)

		job := <-w.workerContractsJobChan
		if job == nil {
			continue
		}
		atomic.SwapInt64(&w.status, InWork)

		w.beforeRun(job.Name())
		job.Run()
		w.afterRun()
	}
}

func (w *worker) beforeRun(name string) {
	w.m.Lock()
	defer w.m.Unlock()
	w.elapsed = 0
	w.jobNameInWork = name
	w.started = time.Now()
}

func (w *worker) afterRun() {
	w.m.Lock()
	defer w.m.Unlock()
	w.elapsed = time.Since(w.started)
}

func (w *worker) getId() int64 {
	w.m.Lock()
	defer w.m.Unlock()
	return w.id
}

func (w *worker) getStatus() int64 {
	return atomic.LoadInt64(&w.status)
}

func (w *worker) metrics() *workerMetric {
	w.m.Lock()
	defer w.m.Unlock()

	return &workerMetric{
		jobName: w.jobNameInWork,
		state:   w.getStatus(),
		started: w.started,
		elapsed: w.elapsed,
	}
}

func (w *worker) close() {
	atomic.SwapInt64(&w.status, Errored)
}

func (w *worker) start(id int64) {
	w.m.Lock()
	defer w.m.Unlock()
	w.id = id
	w.workerContractsJobChan = make(chan Job)

	go w.work()
}
