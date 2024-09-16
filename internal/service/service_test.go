package service

import (
	"testing"

	// model "github.com/FloRichardAloeCorp/gateway/pkg/structs"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/auth"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/ratelimiters"
	"github.com/FloRichardAloeCorp/gateway/internal/test"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type testData struct {
		name            string
		conf            Config
		expectedService *Service
		shouldFail      bool
	}

	provider := test.LaunchTestProvider()

	var testCases = [...]testData{
		{
			name: "Success case",
			conf: Config{
				Name:       "TestService",
				PathPrefix: "/api",
				BaseURL:    "http://localhost:8080",
				Middlewares: ServiceMiddlewares{
					Auth: ServiceAuthConfig{
						Enabled: true,
						AuthMiddlewareConfig: auth.AuthMiddlewareConfig{
							ProviderURL: provider.URL,
							ClientID:    "myapp",
						},
					},
				},
				Endpoints: []EndpointConfiguration{
					{
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			expectedService: &Service{
				name:              "TestService",
				baseURL:           "http://localhost:8080",
				gatewayPathPrefix: "/api",
				authMiddleware:    auth.AuthMiddleware{},
				authEnabled:       true,
				maxBodySize:       0,
				maxHeaderSize:     0,
				endpoints: []EndpointConfiguration{
					{
						Method: "GET",
						Path:   "/test",
						Auth: &EndpointAuth{
							Enabled: true,
						},
					},
				},
			},
			shouldFail: false,
		},
		{
			name: "Fail case: invalid provider url",
			conf: Config{
				Name:       "TestService",
				PathPrefix: "/api",
				BaseURL:    "http://localhost:8080",
				Middlewares: ServiceMiddlewares{
					Auth: ServiceAuthConfig{
						Enabled: true,
						AuthMiddlewareConfig: auth.AuthMiddlewareConfig{
							ProviderURL: "http://invalid.com",
							ClientID:    "myapp",
						},
					},
				},
				Endpoints: []EndpointConfiguration{
					{
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			shouldFail: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service, err := New(testCase.conf)
			if testCase.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedService.name, service.name)
				assert.Equal(t, testCase.expectedService.baseURL, service.baseURL)
				assert.Equal(t, testCase.expectedService.gatewayPathPrefix, service.gatewayPathPrefix)
				assert.Equal(t, testCase.expectedService.authEnabled, service.authEnabled)
				assert.Equal(t, testCase.expectedService.maxBodySize, service.maxBodySize)
				assert.Equal(t, testCase.expectedService.maxHeaderSize, service.maxHeaderSize)
				assert.Equal(t, testCase.expectedService.endpoints, service.endpoints)
			}
		})
	}
}

func TestServiceAttachEndpoints(t *testing.T) {
	type testData struct {
		name    string
		service Service
		router  *gin.Engine
	}

	var testCases = [...]testData{
		{
			name: "Route attached",
			service: Service{
				gatewayPathPrefix: "/api",
				endpoints: []EndpointConfiguration{
					{
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			router: gin.New(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.service.AttachEndpoints(testCase.router)
			for _, route := range testCase.router.Routes() {
				assert.Equal(t, testCase.service.endpoints[0].Method, route.Method)
				assert.Equal(t, testCase.service.gatewayPathPrefix+testCase.service.endpoints[0].Path, route.Path)
			}
		})
	}
}

func TestServiceBuildMiddlewaresChain(t *testing.T) {
	type testData struct {
		name                     string
		serviceConf              Config
		expectedMiddelwaresCount int
	}

	provider := test.LaunchTestProvider()

	var testCases = [...]testData{
		{
			name: "All middlewares activated",
			serviceConf: Config{
				Name:       "TestService",
				PathPrefix: "/api",
				BaseURL:    "http://localhost:8080",
				Middlewares: ServiceMiddlewares{
					Auth: ServiceAuthConfig{
						Enabled: true,
						AuthMiddlewareConfig: auth.AuthMiddlewareConfig{
							ProviderURL: provider.URL,
							ClientID:    "myapp",
						},
					},
					MaxBodySize:   10,
					MaxHeaderSize: 10,
					RateLimit: ServiceRateLimitConfig{
						Enabled: true,
						RateLimiterConfig: ratelimiters.RateLimiterConfig{
							LimitBy:  "sub_claims",
							Window:   10000,
							MaxCount: 5,
						},
					},
				},
				Endpoints: []EndpointConfiguration{
					{
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			expectedMiddelwaresCount: 4,
		},
		{
			name: "All middlewares deactivated",
			serviceConf: Config{
				Name:       "TestService",
				PathPrefix: "/api",
				BaseURL:    "http://localhost:8080",
				Middlewares: ServiceMiddlewares{
					Auth: ServiceAuthConfig{
						Enabled: false,
					},
				},
				Endpoints: []EndpointConfiguration{
					{
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			expectedMiddelwaresCount: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			instance, err := New(testCase.serviceConf)
			assert.NoError(t, err)

			middlewares := instance.buildMiddlewaresChain(instance.endpoints[0])
			assert.Equal(t, testCase.expectedMiddelwaresCount, len(middlewares))
		})
	}
}
