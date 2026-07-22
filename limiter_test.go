package rate-limiter-middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucket_Take(t *testing.T) {
	b := NewTokenBucket(10, 5) // 10 capacity, 5/sec
	// Should allow 10 immediately
	for i := 0; i < 10; i++ {
		if !b.Take(1) {
			t.Fatalf("Take %d failed, should allow", i+1)
		}
	}
	// 11th should fail
	if b.Take(1) {
		t.Fatal("11th Take should fail")
	}
	// Wait for refill
	time.Sleep(200 * time.Millisecond)
	if !b.Take(1) {
		t.Fatal("Should allow after refill")
	}
}

func TestLimiter_Middleware(t *testing.T) {
	l := NewLimiter()
	handler := l.Middleware(5, 10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// First 5 should pass
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
	// 6th should be rate limited
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th request: expected 429, got %d", rec.Code)
	}
}
