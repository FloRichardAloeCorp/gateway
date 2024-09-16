package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Aloe-Corporation/logs"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	log = logs.Get()

	ErrNoAuthHeader          = errors.New("authorization header not found")
	ErrMalformatedAuthHeader = errors.New("authorization header malformated")
)

type AuthMiddlewareConfig struct {
	ProviderURL         string             `mapstructure:"provider_url"`
	ClientID            string             `mapstructure:"client_id"`
	AuthorizedRoles     ClaimCheckerConfig `mapstructure:"authorized_roles"`
	RequiredPermissions ClaimCheckerConfig `mapstructure:"required_permissions"`
}

type AuthMiddleware struct {
	Verifier *oidc.IDTokenVerifier

	roleChecker       *claimChecker
	permissionChecker *claimChecker
}

func NewAuthMiddleware(conf AuthMiddlewareConfig) (*AuthMiddleware, error) {
	provider, err := oidc.NewProvider(context.Background(), conf.ProviderURL)
	if err != nil {
		return nil, fmt.Errorf("can't create new provider: %w", err)
	}

	oidcConfig := oidc.Config{
		ClientID: conf.ClientID,
	}

	verifier := provider.Verifier(&oidcConfig)

	return &AuthMiddleware{
		Verifier:          verifier,
		roleChecker:       newClaimChecker(conf.AuthorizedRoles),
		permissionChecker: newClaimChecker(conf.RequiredPermissions),
	}, nil
}

func (a *AuthMiddleware) Guard(acceptedRoles, acceptedPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, err := extractToken(c)
		if err != nil {
			log.Error("Auth middleware failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, "missing token")
			return
		}

		_, err = a.Verifier.Verify(c.Request.Context(), rawToken)
		if err != nil {
			log.Error("Auth middleware failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token")
			return
		}

		token, _, err := jwt.NewParser().ParseUnverified(rawToken, jwt.MapClaims{})
		if err != nil {
			log.Error("Auth middleware failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token")
			return
		}

		if len(acceptedRoles) > 0 {
			ok, err := a.roleChecker.check(token, acceptedRoles)
			if err != nil {
				log.Error("Auth middleware failure", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid role")
				return
			}

			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid role")
				return
			}
		}

		if len(acceptedPermissions) > 0 {
			ok, err := a.permissionChecker.check(token, acceptedPermissions)
			if err != nil {
				log.Error("Auth middleware failure", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid permission")
				return
			}

			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid role")
				return
			}
		}

		c.Next()
	}
}

func extractToken(c *gin.Context) (string, error) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 {
		return "", ErrMalformatedAuthHeader
	}

	return parts[1], nil
}
