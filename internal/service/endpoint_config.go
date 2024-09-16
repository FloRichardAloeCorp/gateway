package service

import (
	"time"
)

type EndpointConfiguration struct {
	Method        string             `mapstructure:"method"`
	Path          string             `mapstructure:"path"`
	Auth          *EndpointAuth      `mapstructure:"auth,omitempty"`
	RateLimit     *EndpointRateLimit `mapstructure:"rate_limit,omitempty"`
	MaxBodySize   *int64             `mapstructure:"max_body_size,omitempty"`
	MaxHeaderSize *int               `mapstructure:"max_header_size,omitempty"`
}

type EndpointAuth struct {
	// Enable/disable  auth middleware on the endpoint.
	//
	// `service.auth.enabled` must be set to true to enable auth middleware on the endpoint.
	//
	// Set this value to false to disable specific endpoints.
	Enabled            bool     `mapstructure:"enabled"`
	AuthorizedRoles    []string `mapstructure:"authorized_roles"`
	RequiredPermission []string `mapstructure:"required_permissions"`
}

type EndpointRateLimit struct {
	Enabled  bool           `mapstructure:"enabled"`
	LimitBy  *string        `mapstructure:"limit_by"`
	Window   *time.Duration `mapstructure:"window"`
	MaxCount *int           `mapstructure:"max_count"`
}

func (e *EndpointConfiguration) MergeFromServiceConfiguration(conf Config) {
	if e.MaxBodySize == nil && conf.Middlewares.MaxBodySize > 0 {
		e.MaxBodySize = &conf.Middlewares.MaxBodySize
	}

	if e.MaxHeaderSize == nil && conf.Middlewares.MaxHeaderSize > 0 {
		e.MaxHeaderSize = &conf.Middlewares.MaxHeaderSize
	}

	// Injecting whole server rate limit config
	if e.RateLimit == nil && conf.Middlewares.RateLimit.Enabled {
		e.RateLimit = &EndpointRateLimit{
			Enabled:  true,
			LimitBy:  &conf.Middlewares.RateLimit.LimitBy,
			Window:   &conf.Middlewares.RateLimit.Window,
			MaxCount: &conf.Middlewares.RateLimit.MaxCount,
		}
	}

	// Endpoint specifies a rate limit config. Inject missing value from
	// service configuration.
	if e.RateLimit != nil && e.RateLimit.Enabled {
		if e.RateLimit.LimitBy == nil {
			e.RateLimit.LimitBy = &conf.Middlewares.RateLimit.LimitBy
		}

		if e.RateLimit.Window == nil {
			e.RateLimit.Window = &conf.Middlewares.RateLimit.Window
		}

		if e.RateLimit.MaxCount == nil {
			e.RateLimit.MaxCount = &conf.Middlewares.RateLimit.MaxCount
		}
	}

	if e.Auth == nil && conf.Middlewares.Auth.Enabled {
		e.Auth = &EndpointAuth{
			Enabled:            true,
			AuthorizedRoles:    conf.Middlewares.Auth.AuthMiddlewareConfig.AuthorizedRoles.Values,
			RequiredPermission: conf.Middlewares.Auth.AuthMiddlewareConfig.RequiredPermissions.Values,
		}
	}
}
