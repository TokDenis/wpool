package wpool

import (
	"encoding/json"
	"sync"
	"time"
)

type metrics struct {
	throwerStatus Status
	serveStatus   Status
	workers       map[int64]*workerMetric // [workerId]
	sync.Mutex
}

func newMetrics() *metrics {
	return &metrics{
		throwerStatus: 0,
		serveStatus:   0,
		workers:       make(map[int64]*workerMetric),
	}
}

type metricsJson struct {
	ThrowerStatus string              `json:"thrower_status"`
	ServeStatus   string              `json:"serve_status"`
	Workers       []*workerMetricJson `json:"workers"`
}

type workerMetricJson struct {
	Id      int64         `json:"id"`
	JobName string        `json:"job_name"`
	State   string        `json:"state"`
	Started time.Time     `json:"started"`
	Elapsed time.Duration `json:"elapsed"`
}

type Status int

const (
	Unknown Status = iota
	Ok
	Err
	Stop
)

var sts = []string{"Unknown", "Ok", "Err", "Stop"}
var workersStates = []string{"Free", "InWork", "Errored"}

func (s Status) String() string {
	if int(s) > len(sts) {
		return sts[0]
	}
	return sts[s]
}

func (wi *metrics) UpdateWorkerInfo(id int64, m *workerMetric) {
	wi.Lock()
	defer wi.Unlock()
	wi.workers[id] = m
}

func (wi *metrics) SetThrowerStatus(s Status) {
	wi.Lock()
	defer wi.Unlock()
	wi.throwerStatus = s
}

func (wi *metrics) SetServeStatus(s Status) {
	wi.Lock()
	defer wi.Unlock()
	wi.serveStatus = s
}

func (wi *metrics) ToJson() ([]byte, error) {
	wi.Lock()
	defer wi.Unlock()

	var wsMetrics []*workerMetricJson

	for i, metric := range wi.workers {
		wsMetrics = append(wsMetrics, &workerMetricJson{
			Id:      i,
			JobName: metric.jobName,
			State:   workersStates[metric.state],
			Started: metric.started,
			Elapsed: metric.elapsed,
		})
	}

	res := metricsJson{
		ThrowerStatus: wi.throwerStatus.String(),
		ServeStatus:   wi.serveStatus.String(),
		Workers:       wsMetrics,
	}

	return json.Marshal(res)
}
