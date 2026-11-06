package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"backend/internal/platform/apperr"
	"backend/internal/platform/httpx"
	"backend/internal/platform/reqmeta"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter limita las peticiones por IP con un token bucket.
type RateLimiter struct {
	mu          sync.Mutex
	visitors    map[string]*visitor
	limit       rate.Limit
	burst       int
	idleTTL     time.Duration
	lastCleanup time.Time
	now         func() time.Time
}

// NewRateLimiter permite perMinute peticiones por minuto por IP, con ráfagas de burst.
func NewRateLimiter(perMinute, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: map[string]*visitor{},
		limit:    rate.Limit(float64(perMinute) / 60),
		burst:    burst,
		idleTTL:  10 * time.Minute,
		now:      time.Now,
	}
}

// Allow consume un token de la IP y dice si la petición puede seguir.
func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if now.Sub(l.lastCleanup) > time.Minute {
		for k, v := range l.visitors {
			if now.Sub(v.lastSeen) > l.idleTTL {
				delete(l.visitors, k)
			}
		}
		l.lastCleanup = now
	}

	v, ok := l.visitors[key]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.visitors[key] = v
	}
	v.lastSeen = now
	return v.limiter.AllowN(now, 1)
}

// Middleware responde 429 cuando la IP supera el límite.
func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(reqmeta.ClientIP(r.Context())) {
			w.Header().Set("Retry-After", "60")
			httpx.Error(w, r, apperr.ErrTooMany)
			return
		}
		next.ServeHTTP(w, r)
	})
}
