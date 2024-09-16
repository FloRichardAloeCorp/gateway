package ratelimiters

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedWindowCounterAllow(t *testing.T) {
	limiter := &fixedWindowCounter{
		window:   2 * time.Second,
		maxCount: 2,
		counters: make(map[string]*counter),
		mu:       sync.Mutex{},
	}

	key := "id"

	ok := limiter.allow(key)
	assert.True(t, ok)
	ok = limiter.allow(key)
	assert.True(t, ok)
	ok = limiter.allow(key)
	assert.False(t, ok)

	time.Sleep(2 * time.Second)
	ok = limiter.allow(key)
	assert.True(t, ok)
}
