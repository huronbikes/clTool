package clSearch

import (
	"sync"
)

const MaxWorkers = 5

type workerPool struct {
	workerMutex sync.Mutex
	workerCount int
	workerLimiter chan int
}

func (pool *workerPool) init () {
	pool.workerLimiter = make(chan int, MaxWorkers)
}

func (pool *workerPool) AddWorker() {
	pool.workerLimiter <- 1
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	pool.workerCount++
}

func (pool *workerPool) WorkerCount() int{
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	return pool.workerCount
}

func (pool *workerPool) WorkerCompleted() {
	<- pool.workerLimiter
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	pool.workerCount--
}
