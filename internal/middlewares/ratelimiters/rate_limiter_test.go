package ratelimiters

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/FloRichardAloeCorp/gateway/internal/test"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	type testData struct {
		name        string
		conf        RateLimiterConfig
		expectedRes *RateLimiter
	}

	var testCases = [...]testData{
		{
			name: "Success case",
			conf: RateLimiterConfig{
				LimitBy:  "sub_claim",
				Window:   2 * time.Second,
				MaxCount: 2,
			},
			expectedRes: &RateLimiter{
				limiter: fixedWindowCounter{
					window:   2 * time.Second,
					maxCount: 2,
					counters: make(map[string]*counter),
					mu:       sync.Mutex{},
				},
				retrieveLimitingKey: retrieveSubClaim,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			limiter := NewRateLimiter(testCase.conf)
			assert.Equal(t, testCase.expectedRes.limiter.window, limiter.limiter.window)
			assert.Equal(t, testCase.expectedRes.limiter.maxCount, limiter.limiter.maxCount)
		})
	}
}

func TestRateLimiterAllowWithSubClaimRetriever(t *testing.T) {
	header := http.Header{
		"Authorization": []string{
			"Bearer " + test.NewToken(jwt.MapClaims{
				"iss": "issuer",
				"sub": "id",
				"aud": jwt.ClaimStrings{"123456"},
				"exp": jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				"nbf": jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				"iat": jwt.NewNumericDate(time.Now()),
				"jti": "id",
			}),
		},
	}

	rateLimiter := &RateLimiter{
		limiter: fixedWindowCounter{
			window:   2 * time.Second,
			maxCount: 2,
			counters: make(map[string]*counter),
			mu:       sync.Mutex{},
		},
		retrieveLimitingKey: retrieveSubClaim,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}
	c.Request.Header = header
	rateLimiter.Allow()(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = &http.Request{}
	c.Request.Header = header
	rateLimiter.Allow()(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = &http.Request{}
	c.Request.Header = header
	rateLimiter.Allow()(c)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// No token in header

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = &http.Request{}
	rateLimiter.Allow()(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
