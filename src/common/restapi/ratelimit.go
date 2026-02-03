package restapi

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// routeBucket represents a token bucket for any route or endpoint. Tokens should be consumed from a bucket before a request
// is made, and only one thread can actively consume from the bucket at a time.
type routeBucket struct {
	sync.Mutex // built-in mutex is used for ensuring only one thread is consuming at a time
	id         string
	remaining  uint64
	limit      uint64
	reset      time.Time
	mu         sync.Mutex // mu is used for accessing the bucket's contents
}

// consume will block until the bucket has a token available for consumption, then consume it. If true is returned, the
// bucket should only be unlocked after an http response for it has been returned.
func (b *routeBucket) consume() (deferUnlock bool) {
	b.Lock()
	defer func() {
		if !deferUnlock {
			b.Unlock()
		}
	}()

	b.mu.Lock()
	defer b.mu.Unlock()

	wait := b.reset.Sub(time.Now())

	if b.remaining > 0 {
		b.remaining--
		// If the bucket is now empty & does not have a new reset time, unlocking must be deferred until after
		// a response.
		return b.remaining == 0 && wait < 0
	}

	// If reset is in the future, we'll wait until then and provisionally assign ourselves 50% of the tokens from last time.
	// This allows for lower overall latency, but could introduce more 429s. May change in future.
	if wait > 0 {
		b.mu.Unlock() // Temporarily unlock bucket contents while sleeping
		time.Sleep(wait)
		b.mu.Lock()
		b.remaining = b.limit / 2
	}

	b.remaining--
	return false
}

func (b *routeBucket) update(headers http.Header) error {
	hLimit := headers.Get("X-RateLimit-Limit")
	hRemaining := headers.Get("X-RateLimit-Remaining")
	hReset := headers.Get("X-RateLimit-Reset")
	hResetAfter := headers.Get("X-RateLimit-Reset-After")

	b.mu.Lock()
	defer b.mu.Unlock()

	if hLimit != "" {
		limit, err := strconv.ParseUint(hLimit, 10, 64)
		if err != nil {
			return err
		}
		b.limit = limit
	}

	if hRemaining != "" {
		remaining, err := strconv.ParseUint(hRemaining, 10, 64)
		if err != nil {
			return err
		}
		b.remaining = remaining
	}

	// Reset-After seems to be more accurate than Reset so it should be prioritized
	if hResetAfter != "" {
		resetAfter, err := strconv.ParseFloat(hResetAfter, 64)
		if err != nil {
			return err
		}
		b.reset = time.Now().Add(time.Duration(resetAfter * float64(time.Second)))
	} else if hReset != "" {

	}
	return nil
}
