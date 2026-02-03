package restapi

import (
	"math"
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
func (b *routeBucket) consume() {
	b.Lock()

	b.mu.Lock()
	defer b.mu.Unlock()

	// If reset is in the future, we have to wait until then.
	if wait := b.getWaitTime(); wait > 0 {
		time.Sleep(wait)
	}

	b.remaining--
}

// getWaitTime calculates the amount of time the bucket needs to wait to be able to access a token
func (b *routeBucket) getWaitTime() time.Duration {
	if b.remaining > 0 {
		return 0
	}
	return b.reset.Sub(time.Now())
}

func (b *routeBucket) update(headers http.Header) error {
	if headers == nil {
		return nil
	}

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
		reset, err := strconv.ParseFloat(hReset, 64)
		if err != nil {
			return err
		}

		whole, frac := math.Modf(reset)
		b.reset = time.Unix(int64(whole), int64(frac*float64(time.Second)))
	}
	return nil
}
