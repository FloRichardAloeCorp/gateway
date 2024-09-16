package auth

import (
	"fmt"
	"testing"

	// model "github.com/FloRichardAloeCorp/gateway/pkg/structs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestClaimCheckerCheck(t *testing.T) {
	type testData struct {
		name           string
		instance       claimChecker
		token          *jwt.Token
		acceptedValues []string
		shouldFail     bool
		expectedErr    error
		expectedRes    bool
	}
	var testCases = [...]testData{
		{
			name: "Success case: no nesting with string type",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "user",
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: true,
		},
		{
			name: "Success case: no nesting with []string type",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "[]string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": []any{
						"user",
						"dummy",
					},
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: true,
		},
		{
			name: "Success case: nested keys with string type",
			instance: claimChecker{
				tokenKey:  "access.role",
				claimType: "string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"access": map[string]any{
						"role": "user",
					},
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: true,
		},
		{
			name: "Success case: nested keys with []string type",
			instance: claimChecker{
				tokenKey:  "access.role",
				claimType: "[]string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"access": map[string]any{
						"role": []any{
							"user",
							"dummy",
						},
					},
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: true,
		},
		{
			name: "Success case: deeply nested keys with []string type",
			instance: claimChecker{
				tokenKey:  "resource_access.gateway.roles",
				claimType: "[]string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"resource_access": map[string]any{
						"gateway": map[string]any{
							"roles": []any{
								"user",
							},
						},
					},
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: true,
		},
		{
			name: "Success case: invalid role with string type",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "admin",
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: false,
		},
		{
			name: "Success case: invalid role with []string type",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "[]string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": []any{"admin"},
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  false,
			expectedRes: false,
		},
		{
			name: "Fail case: no claims",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "string",
			},
			token: &jwt.Token{},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  true,
			expectedErr: ErrInvalidClaim,
			expectedRes: false,
		},
		{
			name: "Fail case: unknow key",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"invalid_key": "user",
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  true,
			expectedErr: ErrTokenKeyNotFound,
			expectedRes: false,
		},
		{
			name: "Fail case: unsupported token values",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": nil,
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  true,
			expectedErr: fmt.Errorf("can't cast claim to string: %w", ErrInvalidClaimType),
			expectedRes: false,
		},
		{
			name: "Fail case: invalid nested claims",
			instance: claimChecker{
				tokenKey:  "access.role",
				claimType: "string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"access": "role",
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  true,
			expectedErr: ErrInvalidClaim,
			expectedRes: false,
		},
		{
			name: "Fail case: []string type provided but not in token",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "[]string",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "user",
				},
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			shouldFail:  true,
			expectedRes: false,
			expectedErr: fmt.Errorf("can't cast claim to []string: %w", ErrInvalidClaimType),
		},
		{
			name: "Fail case: unsupported claim type",
			instance: claimChecker{
				tokenKey:  "role",
				claimType: "unsupported",
			},
			acceptedValues: []string{
				"user",
				"manager",
			},
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "user",
				},
			},
			shouldFail:  true,
			expectedRes: false,
			expectedErr: ErrUnsupportedClaimType,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			valid, err := testCase.instance.check(testCase.token, testCase.acceptedValues)
			if testCase.shouldFail {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedRes, valid)
			}
		})
	}
}

func TestFindClaim(t *testing.T) {
	type testData struct {
		name          string
		shouldFail    bool
		key           string
		token         *jwt.Token
		expectedClaim any
		expectedErr   error
	}

	var testCases = [...]testData{
		{
			name: "Success case: without nested key",
			key:  "role",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "user",
				},
			},
			shouldFail:    false,
			expectedClaim: "user",
		},
		{
			name: "Success case: with nested key",
			key:  "access.role",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"access": map[string]any{
						"role": "user",
					},
				},
			},
			shouldFail:    false,
			expectedClaim: "user",
		},
		{
			name:        "Fail case: no claims in token",
			key:         "role",
			token:       &jwt.Token{},
			shouldFail:  true,
			expectedErr: ErrInvalidClaim,
		},
		{
			name: "Fail case: unknow key in token",
			key:  "role",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"unknow": "user",
				},
			},
			shouldFail:  true,
			expectedErr: ErrTokenKeyNotFound,
		},
		{
			name: "Fail case: invalid nested value in token",
			key:  "access.role",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"access": nil,
				},
			},
			shouldFail:  true,
			expectedErr: ErrInvalidClaim,
		},
		{
			name: "Fail case: no token key",
			key:  "",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "user",
				},
			},
			shouldFail:  true,
			expectedErr: ErrTokenKeyNotFound,
		},
		{
			name: "Fail case: token key longer than claims",
			key:  "access.role.test",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"role": "user",
				},
			},
			shouldFail:  true,
			expectedErr: ErrTokenKeyNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim, err := findClaim(testCase.key, testCase.token)
			if testCase.shouldFail {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedClaim, claim)
			}
		})
	}
}
