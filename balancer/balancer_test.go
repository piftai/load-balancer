package balancer

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestBackendAliveStatus(t *testing.T) {
	b := &Backend{URL: "http://test", Alive: false}

	if b.isAlive() != false {
		t.Error("Expected backend to be initially dead")
	}

	b.setAlive(true)
	if b.isAlive() != true {
		t.Error("Expected backend to be alive after setAlive(true)")
	}
}

func TestBackendHealthCheck(t *testing.T) {
	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	b := &Backend{URL: ts.URL, Alive: false}
	b.HealthCheck()

	if !b.isAlive() {
		t.Error("Expected backend to be alive after health check")
	}
}

func TestBackendHealthCheckFail(t *testing.T) {
	// Сервер, который возвращает ошибку
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	b := &Backend{URL: ts.URL, Alive: true}
	b.HealthCheck()

	if b.isAlive() {
		t.Error("Expected backend to be dead after failed health check")
	}
}

func TestWRRNext(t *testing.T) {
	backends := []*Backend{
		{URL: "http://backend1", Weight: 3, Alive: true},
		{URL: "http://backend2", Weight: 1, Alive: true},
	}

	wr := New(backends, 10)

	// Проверяем распределение согласно весам
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		backend := wr.Next()
		if backend == nil {
			t.Fatal("Expected non-nil backend")
		}
		counts[backend.URL]++
	}

	// Примерное соотношение должно быть 3:1
	ratio := float64(counts["http://backend1"]) / float64(counts["http://backend2"])
	if ratio < 2.5 || ratio > 3.5 {
		t.Errorf("Expected ratio ~3:1, got %f", ratio)
	}
}

func TestWRRNextWithDeadBackend(t *testing.T) {
	backends := []*Backend{
		{URL: "http://backend1", Weight: 1, Alive: false},
		{URL: "http://backend2", Weight: 1, Alive: true},
	}

	wr := New(backends, 10)

	// Должен всегда возвращать только живой бэкенд
	for i := 0; i < 10; i++ {
		backend := wr.Next()
		if backend == nil {
			t.Fatal("Expected non-nil backend")
		}
		if backend.URL != "http://backend2" {
			t.Errorf("Expected only alive backend, got %s", backend.URL)
		}
	}
}

func TestWRRNextAllDead(t *testing.T) {
	backends := []*Backend{
		{URL: "http://backend1", Weight: 1, Alive: false},
		{URL: "http://backend2", Weight: 1, Alive: false},
	}

	wr := New(backends, 10)

	backend := wr.Next()
	if backend != nil {
		t.Error("Expected nil when all backends are dead")
	}
}

func TestHealthChecker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backends := []*Backend{
		{URL: ts.URL, Weight: 1, Alive: false},
	}

	// Создаем канал для остановки
	stopChan := make(chan struct{})

	// Запускаем HealthChecker в отдельной горутине с возможностью остановки
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var wg sync.WaitGroup
				for _, backend := range backends {
					wg.Add(1)
					go func(b *Backend) {
						defer wg.Done()
						b.HealthCheck()
					}(backend)
				}
				wg.Wait()
			case <-stopChan:
				return
			}
		}
	}()

	// Даем время на выполнение health check
	time.Sleep(50 * time.Millisecond)

	// Останавливаем health checker
	close(stopChan)

	if !backends[0].isAlive() {
		t.Error("Expected backend to be alive after health check")
	}
}

func TestConcurrentAccess(t *testing.T) {
	backends := []*Backend{
		{URL: "http://backend1", Weight: 1, Alive: true},
		{URL: "http://backend2", Weight: 1, Alive: true},
	}

	wr := New(backends, 10)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				backend := wr.Next()
				if backend == nil {
					t.Error("Got nil backend")
				}
			}
		}()
	}
	wg.Wait()
}
