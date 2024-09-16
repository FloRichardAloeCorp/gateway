package ratelimiters

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoAuthHeader          = errors.New("authorization header not found")
	ErrMalformatedAuthHeader = errors.New("authorization header malformated")
	ErrMissingSubClaim       = errors.New("sub claim is missing in token")
)

type KeyRetriever func(c *gin.Context) (string, error)

func selectKeyRetriever(limitBy string) KeyRetriever {
	switch limitBy {
	case "sub_claim":
		return retrieveSubClaim
	default:
		return defaultKeyRetriever
	}
}

var retrieveSubClaim = func(c *gin.Context) (string, error) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 {
		return "", ErrMalformatedAuthHeader
	}

	token, _, err := jwt.NewParser().ParseUnverified(parts[1], jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	if sub == "" {
		return "", ErrMissingSubClaim
	}

	return sub, nil
}

var defaultKeyRetriever = func(c *gin.Context) (string, error) {
	return "global", nil
}
