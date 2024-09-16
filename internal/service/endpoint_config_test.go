package service

import (
	"testing"
	"time"

	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/auth"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/ratelimiters"
	"github.com/stretchr/testify/assert"
)

func TestEndpointConfigurationMergeFromServiceConfigurationMaxBodySize(t *testing.T) {
	type testData struct {
		name           string
		conf           Config
		enpointConfig  EndpointConfiguration
		expectedResult EndpointConfiguration
	}

	var testCases = [...]testData{
		{
			name: "Body size well merged",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					MaxBodySize: 10,
				},
			},
			enpointConfig: EndpointConfiguration{
				MaxBodySize: nil,
			},
			expectedResult: EndpointConfiguration{
				MaxBodySize: int64P(10),
			},
		},
		{
			name: "Service config max body size overriden by endpoint config",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					MaxBodySize: 10,
				},
			},
			enpointConfig: EndpointConfiguration{
				MaxBodySize: int64P(12),
			},
			expectedResult: EndpointConfiguration{
				MaxBodySize: int64P(12),
			},
		},
		{
			name: "Service config max body size set to 0 disable endpoint max body size",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					MaxBodySize: 0,
				},
			},
			enpointConfig: EndpointConfiguration{
				MaxBodySize: nil,
			},
			expectedResult: EndpointConfiguration{
				MaxBodySize: nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.enpointConfig.MergeFromServiceConfiguration(testCase.conf)
			assert.Equal(t, testCase.expectedResult, testCase.enpointConfig)
		})
	}
}

func TestEndpointConfigurationMergeFromServiceConfigurationMaxHeaderSize(t *testing.T) {
	type testData struct {
		name           string
		conf           Config
		enpointConfig  EndpointConfiguration
		expectedResult EndpointConfiguration
	}

	var testCases = [...]testData{
		{
			name: "header size well merged",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					MaxHeaderSize: 10,
				},
			},
			enpointConfig: EndpointConfiguration{
				MaxHeaderSize: nil,
			},
			expectedResult: EndpointConfiguration{
				MaxHeaderSize: intP(10),
			},
		},
		{
			name: "Service config max header size overriden by endpoint config",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					MaxHeaderSize: 10,
				},
			},
			enpointConfig: EndpointConfiguration{
				MaxHeaderSize: intP(12),
			},
			expectedResult: EndpointConfiguration{
				MaxHeaderSize: intP(12),
			},
		},
		{
			name: "Service config max header size set to 0 disable endpoint max header size",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					MaxBodySize: 0,
				},
			},
			enpointConfig: EndpointConfiguration{
				MaxBodySize: nil,
			},
			expectedResult: EndpointConfiguration{
				MaxBodySize: nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.enpointConfig.MergeFromServiceConfiguration(testCase.conf)
			assert.Equal(t, testCase.expectedResult, testCase.enpointConfig)
		})
	}
}

func TestEndpointConfigurationMergeFromServiceConfigurationRateLimit(t *testing.T) {
	type testData struct {
		name           string
		conf           Config
		enpointConfig  EndpointConfiguration
		expectedResult EndpointConfiguration
	}

	oneHourDuration, err := time.ParseDuration("1h")
	assert.NoError(t, err)
	twoHourDuration, err := time.ParseDuration("2h")
	assert.NoError(t, err)

	var testCases = [...]testData{
		{
			name: "Rate limit config well merged",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					RateLimit: ServiceRateLimitConfig{
						Enabled: true,
						RateLimiterConfig: ratelimiters.RateLimiterConfig{
							LimitBy:  "sub_claim",
							Window:   oneHourDuration,
							MaxCount: 10,
						},
					},
				},
			},
			enpointConfig: EndpointConfiguration{
				RateLimit: nil,
			},
			expectedResult: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP("sub_claim"),
					Window:   &oneHourDuration,
					MaxCount: intP(10),
				},
			},
		},
		{
			name: "Rate limit config overriden by endpoint",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					RateLimit: ServiceRateLimitConfig{
						Enabled: true,
						RateLimiterConfig: ratelimiters.RateLimiterConfig{
							LimitBy:  "sub_claim",
							Window:   oneHourDuration,
							MaxCount: 10,
						},
					},
				},
			},
			enpointConfig: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP(""),
					Window:   &twoHourDuration,
					MaxCount: intP(14),
				},
			},
			expectedResult: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP(""),
					Window:   &twoHourDuration,
					MaxCount: intP(14),
				},
			},
		},
		{
			name: "If endpoint config limit by field is nil, it is overridden",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					RateLimit: ServiceRateLimitConfig{
						Enabled: true,
						RateLimiterConfig: ratelimiters.RateLimiterConfig{
							LimitBy:  "sub_claim",
							Window:   oneHourDuration,
							MaxCount: 10,
						},
					},
				},
			},
			enpointConfig: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  nil, // should be overridden
					Window:   &twoHourDuration,
					MaxCount: intP(14),
				},
			},
			expectedResult: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP("sub_claim"),
					Window:   &twoHourDuration,
					MaxCount: intP(14),
				},
			},
		},
		{
			name: "If endpoint config window by field is nil, it is overridden",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					RateLimit: ServiceRateLimitConfig{
						Enabled: true,
						RateLimiterConfig: ratelimiters.RateLimiterConfig{
							LimitBy:  "sub_claim",
							Window:   oneHourDuration,
							MaxCount: 10,
						},
					},
				},
			},
			enpointConfig: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP(""), // should be overridden
					Window:   nil,
					MaxCount: intP(14),
				},
			},
			expectedResult: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP(""),
					Window:   &oneHourDuration,
					MaxCount: intP(14),
				},
			},
		},
		{
			name: "If endpoint config max count by field is nil, it is overridden",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					RateLimit: ServiceRateLimitConfig{
						Enabled: true,
						RateLimiterConfig: ratelimiters.RateLimiterConfig{
							LimitBy:  "sub_claim",
							Window:   oneHourDuration,
							MaxCount: 10,
						},
					},
				},
			},
			enpointConfig: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP(""), // should be overridden
					Window:   &twoHourDuration,
					MaxCount: nil,
				},
			},
			expectedResult: EndpointConfiguration{
				RateLimit: &EndpointRateLimit{
					Enabled:  true,
					LimitBy:  stringP(""),
					Window:   &twoHourDuration,
					MaxCount: intP(10),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.enpointConfig.MergeFromServiceConfiguration(testCase.conf)
			assert.Equal(t, testCase.expectedResult, testCase.enpointConfig)
		})
	}
}

func TestEndpointConfigurationMergeFromServiceConfigurationAuth(t *testing.T) {
	type testData struct {
		name           string
		conf           Config
		enpointConfig  EndpointConfiguration
		expectedResult EndpointConfiguration
	}

	var testCases = [...]testData{
		{
			name: "Auth well merged",
			conf: Config{
				Middlewares: ServiceMiddlewares{
					Auth: ServiceAuthConfig{
						Enabled: true,
						AuthMiddlewareConfig: auth.AuthMiddlewareConfig{
							ProviderURL: "http://provider.com",
							ClientID:    "my",
							AuthorizedRoles: auth.ClaimCheckerConfig{
								TokenKey:  "key",
								ClaimType: "type",
								Values: []string{
									"user",
								},
							},
							RequiredPermissions: auth.ClaimCheckerConfig{
								TokenKey:  "key",
								ClaimType: "type",
								Values: []string{
									"user",
								},
							},
						},
					},
				},
			},
			enpointConfig: EndpointConfiguration{
				Auth: nil,
			},
			expectedResult: EndpointConfiguration{
				Auth: &EndpointAuth{
					Enabled:            true,
					AuthorizedRoles:    []string{"user"},
					RequiredPermission: []string{"user"},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.enpointConfig.MergeFromServiceConfiguration(testCase.conf)
			assert.Equal(t, testCase.expectedResult, testCase.enpointConfig)
		})
	}
}

func int64P(i int64) *int64 {
	return &i
}

func intP(i int) *int {
	return &i
}

func stringP(s string) *string {
	return &s
}
