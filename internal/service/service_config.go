package service

import (
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/auth"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/ratelimiters"
)

type Config struct {
	Name        string                  `mapstructure:"name"`
	PathPrefix  string                  `mapstructure:"path_prefix"`
	BaseURL     string                  `mapstructure:"base_url"`
	Middlewares ServiceMiddlewares      `mapstructure:"middlewares"`
	Endpoints   []EndpointConfiguration `mapstructure:"endpoints"`
}

type ServiceMiddlewares struct {
	Auth          ServiceAuthConfig      `mapstructure:"auth"`
	MaxBodySize   int64                  `mapstructure:"max_body_size"`
	MaxHeaderSize int                    `mapstructure:"max_header_size"`
	RateLimit     ServiceRateLimitConfig `mapstructure:"rate_limit"`
}

type ServiceAuthConfig struct {
	// Enable/disable auth middleware on all endpoints.
	//
	// `middlewares.auth` must be configured to use auth middleware.
	Enabled              bool                      `mapstructure:"enabled"`
	AuthMiddlewareConfig auth.AuthMiddlewareConfig `mapstructure:",omitempty"`
}

type ServiceRateLimitConfig struct {
	Enabled                        bool `mapstructure:"enabled"`
	ratelimiters.RateLimiterConfig `mapstructure:",squash"`
}
