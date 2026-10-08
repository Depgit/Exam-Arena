package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/exam-arena/internal/platform/respond"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64
	burst    int
}

func newRateLimiter(rps float64, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rps,
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:   float64(rl.burst) - 1,
			lastSeen: time.Now(),
		}
		return true
	}

	elapsed := time.Since(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.lastSeen = time.Now()

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func RateLimit(rps float64, burst int, next http.Handler) http.Handler {
	limiter := newRateLimiter(rps, burst)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		if !limiter.allow(ip) {
			respond.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}
