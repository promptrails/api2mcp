package engine

import (
	"sync"
	"time"
)

// RateLimit caps how often a single tool may be invoked. It is applied
// per-operation (each tool gets its own bucket), so one chatty tool can't
// starve the others. A zero PerSecond disables limiting.
type RateLimit struct {
	// PerSecond is the sustained refill rate (tokens added per second).
	PerSecond float64
	// Burst is the bucket capacity — the most calls allowed back-to-back.
	Burst int
}

func (r RateLimit) enabled() bool { return r.PerSecond > 0 && r.Burst > 0 }

// tokenBucket is a minimal, dependency-free token-bucket limiter. allow() is
// non-blocking: it returns false immediately when no token is available, so a
// rate-limited tool call fails fast with a clear message instead of hanging the
// LLM.
type tokenBucket struct {
	mu     sync.Mutex
	tokens float64
	max    float64
	refill float64 // tokens per second
	last   time.Time
	nowFn  func() time.Time
}

func newBucket(r RateLimit) *tokenBucket {
	return &tokenBucket{
		tokens: float64(r.Burst),
		max:    float64(r.Burst),
		refill: r.PerSecond,
		last:   time.Now(),
		nowFn:  time.Now,
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.nowFn()
	b.tokens += now.Sub(b.last).Seconds() * b.refill
	if b.tokens > b.max {
		b.tokens = b.max
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}
