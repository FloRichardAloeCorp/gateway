package auth

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidClaim     = errors.New("invalid claim")
	ErrTokenKeyNotFound = errors.New("provided token key not found in token")

	ErrInvalidClaimType     = errors.New("can't cast token claim to provided type")
	ErrUnsupportedClaimType = errors.New("unsupported provided claim type")
)

type ClaimCheckerConfig struct {
	TokenKey  string   `mapstructure:"token_key"`
	ClaimType string   `mapstructure:"claim_type"`
	Values    []string `mapstructure:"values"`
}

type claimChecker struct {
	tokenKey  string
	claimType string
}

func newClaimChecker(conf ClaimCheckerConfig) *claimChecker {
	return &claimChecker{
		tokenKey:  conf.TokenKey,
		claimType: conf.ClaimType,
	}
}

func (c *claimChecker) check(token *jwt.Token, acceptedValues []string) (bool, error) {
	rawClaim, err := findClaim(c.tokenKey, token)
	if err != nil {
		return false, err
	}

	switch c.claimType {
	case "string":
		claim, ok := rawClaim.(string)
		if !ok {
			return false, fmt.Errorf("can't cast claim to string: %w", ErrInvalidClaimType)
		}
		return slices.Contains(acceptedValues, claim), nil
	case "[]string":
		claim, ok := rawClaim.([]any)
		if !ok {
			return false, fmt.Errorf("can't cast claim to []string: %w", ErrInvalidClaimType)
		}

		for _, value := range claim {
			strValue, ok := value.(string)
			if !ok {
				return false, fmt.Errorf("can't cast claim element to string: %w", ErrInvalidClaimType)
			}

			if slices.Contains(acceptedValues, strValue) {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, ErrUnsupportedClaimType
	}
}

func findClaim(key string, token *jwt.Token) (any, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaim
	}

	parts := strings.Split(key, ".")
	for i, part := range parts {
		claim, ok := claims[part]
		if !ok {
			return false, ErrTokenKeyNotFound
		}

		if i == len(parts)-1 {
			return claim, nil
		}

		claims, ok = claim.(map[string]any)
		if !ok {
			return false, ErrInvalidClaim
		}
	}

	return nil, ErrTokenKeyNotFound
}
