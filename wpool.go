package wpool

import (
	"context"
	"time"
)

type WPool interface {
	InfoJson() ([]byte, error)
	ConsumeJob(j Job)
	JobChan() chan Job
}

type wPool struct {
	jobChan chan Job
	workers []*worker

	info *metrics

	opt Options
}

func (wp *wPool) InfoJson() ([]byte, error) {
	for _, w := range wp.workers {
		wp.info.UpdateWorkerInfo(w.getId(), w.metrics())
	}

	return wp.info.ToJson()
}

func (wp *wPool) ConsumeJob(j Job) {
	wp.jobChan <- j
}

func (wp *wPool) JobChan() chan Job {
	return wp.jobChan
}

func (wp *wPool) start(ctx context.Context) {
	go wp.serve(ctx)
	go wp.thrower(ctx)
}

func (wp *wPool) serve(ctx context.Context) {
	wp.info.SetServeStatus(Ok)
	for {
		select {
		case <-ctx.Done():
			wp.info.SetServeStatus(Stop)
			return
		default:
			for _, w := range wp.workers {
				if w.getStatus() == Errored {
					w.start(w.id)
				}
			}
		}
		time.Sleep(time.Millisecond)
	}

}

func (wp *wPool) thrower(ctx context.Context) {
	wp.info.SetThrowerStatus(Ok)
	for {
		for _, w := range wp.workers {
			if w.getStatus() == Free {
				select {
				case job := <-wp.jobChan:
					w.consume(job)
				case <-ctx.Done():
					wp.info.SetThrowerStatus(Stop)
					return
				}
			}
		}
	}
}

func New(ctx context.Context, opt Options) WPool {
	workers := make([]*worker, opt.MaxWorker)

	for i := 0; i < len(workers); i++ {
		w := worker{}
		w.start(int64(i))
		workers[i] = &w
	}

	p := wPool{
		jobChan: make(chan Job, opt.JobBuffer),
		workers: workers,
		opt:     opt,
		info:    newMetrics(),
	}

	p.start(ctx)

	return &p
}
