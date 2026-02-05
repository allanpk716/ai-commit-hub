// Package concurrency provides utilities for managing concurrent operations
package concurrency

import (
	"context"
	"runtime"
	"sync"
)

// WorkerPool 并发工作池
type WorkerPool struct {
	maxWorkers int
	wg         sync.WaitGroup
	sem        chan struct{}
}

// NewWorkerPool 创建工作池
func NewWorkerPool(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	return &WorkerPool{
		maxWorkers: maxWorkers,
		sem:        make(chan struct{}, maxWorkers),
	}
}

// Submit 提交任务到工作池
func (p *WorkerPool) Submit(ctx context.Context, fn func() error) error {
	select {
	case p.sem <- struct{}{}:
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			defer func() { <-p.sem }()
			_ = fn()
		}()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wait 等待所有任务完成
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// DynamicConcurrency 根据负载动态调整并发数
func DynamicConcurrency(minItems, maxConcurrency int) int {
	if minItems < maxConcurrency {
		return minItems
	}

	// 根据 CPU 核心数动态调整
	cpuCount := runtime.NumCPU()
	if cpuCount < 4 {
		return min(5, maxConcurrency)
	}

	return maxConcurrency
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
