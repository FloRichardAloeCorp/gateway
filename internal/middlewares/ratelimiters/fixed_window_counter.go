package ratelimiters

import (
	"sync"
	"time"
)

type FixedWindowCounterConf struct {
	Window string `mapstructure:"window"`
}

type fixedWindowCounter struct {
	window   time.Duration
	maxCount int
	counters map[string]*counter
	mu       sync.Mutex
}

type counter struct {
	timestamp time.Time
	count     int
}

func (f *fixedWindowCounter) allow(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	window, ok := f.counters[key]
	if !ok || time.Since(window.timestamp) > f.window {
		f.counters[key] = &counter{
			timestamp: time.Now().UTC(),
			count:     1,
		}
		return true
	}

	if window.count < f.maxCount {
		window.count++
		return true
	}

	return false
}
