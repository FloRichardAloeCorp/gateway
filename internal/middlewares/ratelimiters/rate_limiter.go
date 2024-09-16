package ratelimiters

import (
	"net/http"
	"sync"
	"time"

	"github.com/Aloe-Corporation/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	log = logs.Get()
)

type RateLimiterConfig struct {
	LimitBy  string        `mapstructure:"limit_by"`
	Window   time.Duration `mapstructure:"window"`
	MaxCount int           `mapstructure:"max_count"`
}

type RateLimiter struct {
	limiter             fixedWindowCounter
	retrieveLimitingKey KeyRetriever
}

func NewRateLimiter(conf RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		limiter: fixedWindowCounter{
			window:   conf.Window,
			maxCount: conf.MaxCount,
			counters: make(map[string]*counter),
			mu:       sync.Mutex{},
		},
		retrieveLimitingKey: selectKeyRetriever(conf.LimitBy),
	}
}

func (r *RateLimiter) Allow() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := r.retrieveLimitingKey(c)
		if err != nil {
			log.Error("RateLimiter middleware failure", zap.Error(err))
			c.JSON(http.StatusBadRequest, "Bad Request")
			return
		}

		if !r.limiter.allow(key) {
			log.Error("RateLimiter middleware blocking", zap.String("reason", "rate limit exceeded"))
			c.JSON(http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		c.Next()
	}
}
