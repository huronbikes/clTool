package clSearch

import (
	"sync"
)

const MaxWorkers = 5

type workerPool struct {
	workerMutex sync.Mutex
	count int
	workerLimiter chan int
	closed bool
}

func (pool *workerPool) init () {
	pool.workerLimiter = make(chan int, MaxWorkers)
}

func (pool *workerPool) stop() {
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	pool.closed=true
}

func (pool *workerPool) stopped () bool {
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	return pool.closed
}

func (pool *workerPool) addWorker(fn func()) {
	if !pool.closed {
		pool.workerLimiter <- 1
		pool.workerMutex.Lock()
		defer pool.workerMutex.Unlock()
		pool.count++
		go fn()
	}
}

func (pool *workerPool) workerCount() int{
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	return pool.count
}

func (pool *workerPool) workerCompleted() {
	<- pool.workerLimiter
	pool.workerMutex.Lock()
	defer pool.workerMutex.Unlock()
	pool.count--
}
