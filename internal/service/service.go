package service

import (
	"fmt"

	"github.com/Aloe-Corporation/logs"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/auth"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/bodysizelimiter"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/headersizelimiter"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/ratelimiters"

	"github.com/FloRichardAloeCorp/gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log = logs.Get()

type Service struct {
	name              string
	baseURL           string
	gatewayPathPrefix string

	authMiddleware auth.AuthMiddleware
	authEnabled    bool

	maxBodySize   int64
	maxHeaderSize int

	endpoints []EndpointConfiguration
}

func New(conf Config) (*Service, error) {
	service := &Service{
		name:              conf.Name,
		baseURL:           conf.BaseURL,
		gatewayPathPrefix: conf.PathPrefix,

		authEnabled: conf.Middlewares.Auth.Enabled,

		maxBodySize:   conf.Middlewares.MaxBodySize,
		maxHeaderSize: conf.Middlewares.MaxHeaderSize,
	}

	mergedEndpoints := []EndpointConfiguration{}
	for _, endpoint := range conf.Endpoints {
		endpoint.MergeFromServiceConfiguration(conf)
		mergedEndpoints = append(mergedEndpoints, endpoint)
	}
	service.endpoints = mergedEndpoints

	if service.authEnabled {
		authMiddleware, err := auth.NewAuthMiddleware(conf.Middlewares.Auth.AuthMiddlewareConfig)
		if err != nil {
			return nil, err
		}

		service.authMiddleware = *authMiddleware
	}

	return service, nil
}

func (s *Service) AttachEndpoints(router *gin.Engine) {
	for _, endpoint := range s.endpoints {
		middlewares := s.buildMiddlewaresChain(endpoint)

		handlers := []gin.HandlerFunc{}
		handlers = append(handlers, middlewares...)
		handlers = append(handlers, proxy.Forward(s.gatewayPathPrefix, s.baseURL))

		router.Handle(endpoint.Method, s.gatewayPathPrefix+endpoint.Path, handlers...)
	}
}

func (s *Service) buildMiddlewaresChain(endpoint EndpointConfiguration) []gin.HandlerFunc {
	handlers := []gin.HandlerFunc{}
	if s.authEnabled && endpoint.Auth.Enabled {
		handlers = append(handlers, s.authMiddleware.Guard(endpoint.Auth.AuthorizedRoles, endpoint.Auth.RequiredPermission))
		log.Info("authorization middleware enabled",
			zap.String("service", s.name),
			zap.String("endpoint", endpoint.Method+" "+endpoint.Path),
		)
	} else {
		log.Warn("path not protected by auth middleware",
			zap.String("service", s.name),
			zap.String("endpoint", endpoint.Method+" "+endpoint.Path),
		)
	}

	if endpoint.MaxBodySize != nil {
		handlers = append(handlers, bodysizelimiter.Limit(*endpoint.MaxBodySize))
		log.Info("request body size middleware enabled",
			zap.Int64p("limit (in bytes)", endpoint.MaxBodySize),
			zap.String("service", s.name),
			zap.String("endpoint", endpoint.Method+" "+endpoint.Path),
		)
	}

	if endpoint.MaxHeaderSize != nil {
		handlers = append(handlers, headersizelimiter.Limit(*endpoint.MaxHeaderSize))
		log.Info("header size middleware enabled",
			zap.Intp("limit (in bytes)", endpoint.MaxHeaderSize),
			zap.String("service", s.name),
			zap.String("endpoint", endpoint.Method+" "+endpoint.Path),
		)
	}

	if endpoint.RateLimit != nil && endpoint.RateLimit.Enabled {
		limiter := ratelimiters.NewRateLimiter(ratelimiters.RateLimiterConfig{
			LimitBy:  *endpoint.RateLimit.LimitBy,
			Window:   *endpoint.RateLimit.Window,
			MaxCount: *endpoint.RateLimit.MaxCount,
		})
		handlers = append(handlers, limiter.Allow())
		log.Info("rate limiter middleware enabled",
			zap.String("limit", fmt.Sprintf("%d requests every %s", *endpoint.RateLimit.MaxCount, endpoint.RateLimit.Window.String())),
			zap.String("service", s.name),
			zap.String("endpoint", endpoint.Method+" "+endpoint.Path),
		)
	}

	return handlers
}
