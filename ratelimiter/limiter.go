package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int           // максимальное кол-во токенов в одном бакете
	tokens     int           // текущее кол-во токенов
	rate       time.Duration // интервал между добавлениями токенов
	lastUpdate time.Time     // время последнего обновления бакета
	mu         sync.Mutex
}

func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // начинаем с полного бакета
		rate:       rate,
		lastUpdate: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	elapsed := time.Since(tb.lastUpdate).Seconds()
	tokensToAdd := int(elapsed * tb.rate.Seconds())

	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity) // если кол-во токенов которое нужно добавить больше, чем capacity, то добавим просто размер capacity
		tb.lastUpdate = time.Now()
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

type ClientLimiter struct {
	buckets         map[string]*TokenBucket
	defaultCapacity int
	defaultRate     time.Duration
	mu              sync.RWMutex
	stopChan        chan struct{}
}

func (c *ClientLimiter) refillBuckets() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			for _, bucket := range c.buckets {
				bucket.mu.Lock()
				elapsed := time.Since(bucket.lastUpdate).Seconds()
				tokensToAdd := int(elapsed * float64(bucket.rate))
				if tokensToAdd > 0 {
					bucket.tokens = min(c.defaultCapacity, tokensToAdd) // если tokensToAdd больше cap, то заполним просто по максимуму
					bucket.lastUpdate = time.Now()
				}
				bucket.mu.Unlock()
			}
			c.mu.RUnlock()
		}
	}
}

func NewClientLimiter(defaultCapacity int, defaultRate time.Duration) *ClientLimiter {
	c := ClientLimiter{
		buckets:         make(map[string]*TokenBucket), // todo может стоит добавить по дефолту для 10 клиентов
		defaultCapacity: defaultCapacity,
		defaultRate:     defaultRate,
		stopChan:        make(chan struct{}),
	}
	go c.refillBuckets()
	return &c
}

func (c *ClientLimiter) Allow(clientID string) bool {
	c.mu.RLock()
	bucket, exists := c.buckets[clientID]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		bucket = NewTokenBucket(c.defaultCapacity, c.defaultRate)
		c.buckets[clientID] = bucket
		c.mu.Unlock()
	}

	return bucket.Allow()
}
