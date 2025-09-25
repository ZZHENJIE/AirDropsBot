package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Task 定时任务函数类型，在每个定时周期执行
type Task func(ctx context.Context)

// Scheduler 调度器，提供定时任务的启动、停止和管理功能
type Scheduler struct {
	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// New 创建新的调度器实例
func New() *Scheduler {
	return &Scheduler{}
}

// Start 按指定的时间间隔启动调度器。如果已在运行，则返回错误
func (s *Scheduler) Start(interval time.Duration, task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return errors.New("scheduler already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		// Run immediately once
		task(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				task(ctx)
			}
		}
	}()
	return nil
}

// Stop stops the scheduler if running.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	cancel := s.cancel
	s.running = false
	s.cancel = nil
	s.mu.Unlock()

	cancel()
	s.wg.Wait()
}

// IsRunning reports whether the scheduler is running.
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
