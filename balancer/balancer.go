// Package balancer отвечает за балансировщик
package balancer

import (
	"sync"
)

type Backend struct {
	URL    string
	Weight int
	Alive  bool
}

type WRR struct {
	backends []Backend
	mu       sync.RWMutex
	index    int // - индекс текущего бекенда
	current  int // - потраченный вес
	stopChan chan struct{}
}

func New(backends []Backend) *WRR {
	w := &WRR{
		backends: backends,
		stopChan: make(chan struct{}),
	}
	return w
}

func (wr *WRR) Next() *Backend {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	for i := 0; i < len(wr.backends); i++ {
		backend := &wr.backends[wr.index]
		wr.current++

		if wr.current >= backend.Weight { // проверяем, истратили ли мы вес текущего бекенда, если да, то переходим на следующий
			wr.current = 0
			wr.index = (wr.index + 1) % len(wr.backends)
		}

		if backend.Alive { // если жив, возвращаем сразу же его
			return backend
		}

		wr.index = (wr.index + 1) % len(wr.backends)
	}

	return nil
}
