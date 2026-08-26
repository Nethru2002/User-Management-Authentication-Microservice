package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientLimiter struct {
	tokens     float64
	lastSeen   time.Time
	maxTokens  float64
	refillRate float64
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimiter
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
	}
	go rl.cleanupRoutine()
	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		rl.mu.Lock()
		limiter, exists := rl.clients[ip]
		now := time.Now()
		if !exists {
			limiter = &clientLimiter{
				tokens:     20,
				maxTokens:  20,
				refillRate: 10,
				lastSeen:   now,
			}
			rl.clients[ip] = limiter
		}

		elapsed := now.Sub(limiter.lastSeen).Seconds()
		limiter.lastSeen = now
		limiter.tokens += elapsed * limiter.refillRate
		if limiter.tokens > limiter.maxTokens {
			limiter.tokens = limiter.maxTokens
		}

		if limiter.tokens < 1 {
			rl.mu.Unlock()
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		limiter.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, limiter := range rl.clients {
			if time.Since(limiter.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}