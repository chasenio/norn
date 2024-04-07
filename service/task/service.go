package task

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"sync"
)

type Service struct {
	queue chan int
	wg    sync.WaitGroup
}

func NewService(lc fx.Lifecycle) *Service {
	s := &Service{
		queue: make(chan int),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			s.Stop()
			return nil
		},
	})

	return s
}

// Start service
func (s *Service) Start() {
	go s.Run() // run task goroutine
	logrus.Infof("task service is running...")
}

// Push task to queue
func (s *Service) Push(task int) {
	s.queue <- task
}

// Run task
func (s *Service) Run() {
	for task := range s.queue {
		s.wg.Add(1)
		go s.Do(task)
	}
}

// Stop service
func (s *Service) Stop() {
	logrus.Infof("receive stop request, stop task service...")
	s.Wait()
}

// Wait for all task done
func (s *Service) Wait() {
	// check task number
	if len(s.queue) != 0 {
		logrus.Infof("queue has %d task(s), waiting for task done.", len(s.queue))
		s.wg.Wait()
	}
	logrus.Infof("all task done.")
	close(s.queue)
}

// Do task
func (s *Service) Do(task int) {
	defer s.wg.Done()
	logrus.Infof("execute task: %d", task)
	// 处理任务逻辑
}
