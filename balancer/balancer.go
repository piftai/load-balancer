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
	mu     sync.RWMutex
}

func (b *Backend) isAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

func (b *Backend) setAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

func (b *Backend) HealthCheck() {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(b.URL) // будем считать, что запрос по root url равноценен запросу /health
	if err != nil {
		log.Printf("Health check. Backend %v not answering", b.URL)
		b.setAlive(false)
		return
	}
	defer resp.Body.Close()

	// Считаем бэкенд живым, если статус код 2xx или 3xx
	b.setAlive(resp.StatusCode < 400)
}

// WRR - Weighted Round Robin - взвешенный алгоритм Round Robin(в IDEAS.MD вынес почему выбрал)
type WRR struct {
	backends []*Backend
	mu       sync.RWMutex
	index    int           // - индекс текущего бекенда
	current  int           // - потраченный вес
	stopChan chan struct{} // для graceful shutdown
	GCD      int           // Greatest Common Divisor для нормализации веса
}

func (wr *WRR) HealthChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var wg sync.WaitGroup
			// todo ограничить число горутин
			for _, backend := range wr.backends {
				wg.Add(1)
				go func(b *Backend) {
					defer wg.Done()
					b.HealthCheck()
					fmt.Printf("Backend %s is alive: %v\n", b.URL, b.isAlive())
				}(backend)
			}

			wg.Wait()
		case <-wr.stopChan:
			return
		}
	}
}

func (wr *WRR) Stop() {
	close(wr.stopChan) // Останавливаем health checker
}

func New(backends []*Backend, healthInterval int) *WRR {
	w := &WRR{
		backends: backends,
		stopChan: make(chan struct{}),
	}

	w.calculateGCD()
	go w.HealthChecker(time.Duration(healthInterval) * time.Second)

	return w
}

func (wr *WRR) calculateGCD() {
	gcd := wr.backends[0].Weight
	for _, backend := range wr.backends {
		gcd = gcdTwoNumbers(gcd, backend.Weight)
	}
	wr.GCD = gcd
}

func gcdTwoNumbers(a, b int) int { // gcd - greatest common divisor
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func (wr *WRR) Next() *Backend {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	for i := 0; i < len(wr.backends); i++ {
		backend := wr.backends[wr.index]
		wr.current++

		if wr.current >= (backend.Weight / wr.GCD) { // проверяем, истратили ли мы вес текущего бекенда, если да, то переходим на следующий
			wr.current = 0                               // обнуляем вес для следующего бекенда
			wr.index = (wr.index + 1) % len(wr.backends) // берем следующий бекенд
		}

		if backend.isAlive() { // если жив, возвращаем сразу же его
			return backend
		}

		wr.index = (wr.index + 1) % len(wr.backends) // если бекенд не отвечает, то сразу же идем на другой, а потом снова проверим его
	}

	return nil
}
