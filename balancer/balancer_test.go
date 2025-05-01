package balancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWRRNext(t *testing.T) {
	tests := []struct {
		name             string
		backends         []Backend
		expectedSequence []string // Ожидаемая последовательность URL (или "nil" для dead backends)
	}{
		{
			name: "equal weights",
			backends: []Backend{
				{URL: "server1", Weight: 1, Alive: true},
				{URL: "server2", Weight: 1, Alive: true},
				{URL: "server3", Weight: 1, Alive: true},
			},
			expectedSequence: []string{
				"server1", "server2", "server3",
				"server1", "server2", "server3", // Цикл повторяется
			},
		},
		{
			name: "different weights",
			backends: []Backend{
				{URL: "server1", Weight: 3, Alive: true},
				{URL: "server2", Weight: 1, Alive: true},
				{URL: "server3", Weight: 2, Alive: true},
			},
			expectedSequence: []string{
				"server1", "server1", "server1", // Вес 3
				"server2",            // Вес 1
				"server3", "server3", // Вес 2
				"server1", "server1", "server1", // Цикл повторяется
				"server2",
				"server3", "server3",
			},
		},
		{
			name: "with dead backends",
			backends: []Backend{
				{URL: "server1", Weight: 1, Alive: false},
				{URL: "server2", Weight: 2, Alive: true},
				{URL: "server3", Weight: 1, Alive: true},
			},
			expectedSequence: []string{
				"server2", "server2", "server3", // server1 пропускается
				"server2", "server2", "server3", // Цикл повторяется
			},
		},
		{
			name: "all dead backends",
			backends: []Backend{
				{URL: "server1", Weight: 1, Alive: false},
				{URL: "server2", Weight: 1, Alive: false},
			},
			expectedSequence: []string{"nil", "nil"}, // Ожидаем nil для всех вызовов
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := New(tt.backends)
			for _, expectedURL := range tt.expectedSequence {
				got := wr.Next()

				if expectedURL == "nil" {
					assert.Nil(t, got, "Expected nil for dead backend")
				} else {
					assert.NotNil(t, got, "Backend should not be nil")
					if got != nil {
						assert.Equal(t, expectedURL, got.URL, "Unexpected backend URL")
					}
				}
			}
		})
	}
}

func TestWRRNoBackends(t *testing.T) {
	wr := New([]Backend{})
	assert.Nil(t, wr.Next(), "Should return nil when no backends available")
}

func TestWRRConcurrency(t *testing.T) {
	backends := []Backend{
		{URL: "server1", Weight: 2, Alive: true},
		{URL: "server2", Weight: 1, Alive: true},
	}
	wr := New(backends)

	results := make(chan *Backend, 100)

	// Запускаем несколько горутин
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				results <- wr.Next()
			}
		}()
	}

	// Проверяем результаты
	for i := 0; i < 100; i++ {
		b := <-results
		assert.NotNil(t, b)
		if b != nil {
			assert.True(t, b.URL == "server1" || b.URL == "server2")
		}
	}
}
