// Package balancer отвечает за балансировщик
package balancer

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Backend struct {
	URL    string
	Weight int
	Alive  bool
}

func (b *Backend) HealthCheck() {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(b.URL) // будем считать, что запрос по root url равноценен запросу /health
	if err != nil {
		log.Printf("Health check. Backend %v not answering", b.URL)
		b.Alive = false
		return
	}
	defer resp.Body.Close()

	// Считаем бэкенд живым, если статус код 2xx или 3xx
	b.Alive = resp.StatusCode < 400
}

func HealthChecker(backends []*Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		var wg sync.WaitGroup

		for _, backend := range backends {
			wg.Add(1)
			go func(b *Backend) {
				defer wg.Done()
				b.HealthCheck()
				fmt.Printf("Backend %s is alive: %v\n", b.URL, b.Alive)
			}(backend)
		}

		wg.Wait()

	}
}

type WRR struct {
	backends []*Backend
	mu       sync.RWMutex
	index    int           // - индекс текущего бекенда
	current  int           // - потраченный вес
	stopChan chan struct{} // для graceful shutdown
}

func New(backends []*Backend, healthInterval int) *WRR {
	w := &WRR{
		backends: backends,
		stopChan: make(chan struct{}),
	}

	go HealthChecker(w.backends, time.Duration(healthInterval)*time.Second)

	return w
}

func (wr *WRR) Next() *Backend {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	for i := 0; i < len(wr.backends); i++ {
		backend := wr.backends[wr.index]
		wr.current++

		if wr.current >= backend.Weight { // проверяем, истратили ли мы вес текущего бекенда, если да, то переходим на следующий
			wr.current = 0                               // обнуляем вес для следующего бекенда
			wr.index = (wr.index + 1) % len(wr.backends) // берем следующий бекенд
		}

		if backend.Alive { // если жив, возвращаем сразу же его
			return backend
		}

		wr.index = (wr.index + 1) % len(wr.backends) // если бекенд не отвечает, то сразу же идем на другой, а потом снова проверим его
	}

	return nil
}
