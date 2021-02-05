package wpool

type Job interface {
	Run()
	Name() string
}
