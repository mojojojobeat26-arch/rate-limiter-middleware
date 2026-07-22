package rate-limiter-middleware

import (
	"context"
	"sync"
	"time"
	"net/http"
)

// TokenBucket implements the token bucket algorithm for rate limiting
// Allows burst traffic up to bucket size while maintaining average rate
type TokenBucket struct {
	mu       sync.Mutex
	capacity int
	tokens   float64
	rate     float64 // tokens per second
	last     time.Time
}

func NewTokenBucket(capacity int, rate float64) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   float64(capacity),
		rate:     rate,
		last:     time.Now(),
	}
}

// Take attempts to consume n tokens, returns true if successful
func (tb *TokenBucket) Take(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.last).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	tb.last = now

	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return true
	}
	return false
}

// TakeWithTimeout blocks until tokens are available or timeout
type Limiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

func NewLimiter() *Limiter {
	return &Limiter{buckets: make(map[string]*TokenBucket)}
}

func (l *Limiter) GetBucket(key string, capacity int, rate float64) *TokenBucket {
	l.mu.RLock()
	b, ok := l.buckets[key]
	l.mu.RUnlock()
	if ok {
		return b
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok = l.buckets[key]; ok {
		return b
	}
	b = NewTokenBucket(capacity, rate)
	l.buckets[key] = b
	return b
}

// Middleware returns an HTTP middleware that rate limits by IP
func (l *Limiter) Middleware(capacity int, rate float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}
			bucket := l.GetBucket(ip, capacity, rate)
			if !bucket.Take(1) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
