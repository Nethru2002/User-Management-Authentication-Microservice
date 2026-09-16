package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"auth-service/pkg/utils"

	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local current = tonumber(redis.call('get', key) or "0")

if current + 1 > limit then
    return 0
else
    redis.call("INCRBY", key, 1)
    if current == 0 then
        redis.call("EXPIRE", key, ARGV[2])
    end
    return 1
end
`)

type RateLimiter struct {
	rdb        *redis.Client
	limit      int
	windowSecs int
}

func NewRedisRateLimiter(rdb *redis.Client, requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		rdb:        rdb,
		limit:      requestsPerMinute,
		windowSecs: 60,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractClientIP(r)
		key := "ratelimit:" + ip

		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		allowed, err := rateLimitScript.Run(ctx, rl.rdb, []string{key}, rl.limit, rl.windowSecs).Int()
		if err == nil && allowed == 0 {
			utils.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded: please try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}