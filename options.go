package wpool

type Options struct {
	JobBuffer int
	MaxWorker int
}

func DefaultOptions() Options {
	return Options{
		JobBuffer: 5,
		MaxWorker: 10,
	}
}

func (opt Options) WithJobBuffer(val int) Options {
	opt.JobBuffer = val
	return opt
}

func (opt Options) WithMaxWorker(val int) Options {
	opt.MaxWorker = val
	return opt
}
