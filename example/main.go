package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wpool"
)

var wp wpool.WPool

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case s := <-c:
			fmt.Printf("signal %v", s)
			cancel()
		case <-ctx.Done():
		}
	}()

	wp = wpool.New(ctx, wpool.DefaultOptions().WithMaxWorker(1000).WithJobBuffer(500))

	//go func() {
	//	for {
	//		b, err := wp.InfoJson()
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//		fmt.Println(string(b))
	//		time.Sleep(time.Second * 10)
	//	}
	//}()

	for i := 0; i < 3000; i++ {
		job := LongJob{}
		job2 := LongJob2{}

		wp.ConsumeJob(&job)
		wp.ConsumeJob(&job2)
	}

	<-ctx.Done()

	time.Sleep(time.Second * 1)
}

type LongJob struct {
}

func (l *LongJob) Run() {
	time.Sleep(time.Second * 2)
	fmt.Println("LongJob done")
}

func (l *LongJob) Name() string {
	return "LongJob"
}

type LongJob2 struct {
}

func (l *LongJob2) Run() {
	time.Sleep(time.Second)
	fmt.Println("LongJob2 done")
}

func (l *LongJob2) Name() string {
	return "LongJob2"
}
