package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// model "github.com/FloRichardAloeCorp/gateway/pkg/structs"

	"github.com/FloRichardAloeCorp/gateway/internal/test"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	type testData struct {
		name       string
		shouldFail bool

		conf        AuthMiddlewareConfig
		expectedErr error
	}

	server := test.LaunchTestProvider()

	var testCases = [...]testData{
		{
			name:       "Success case",
			shouldFail: false,
			conf: AuthMiddlewareConfig{
				ProviderURL: server.URL,
				ClientID:    "1234567890",
			},
		},
		{
			name:       "Fail case: invalid test provider",
			shouldFail: true,
			conf: AuthMiddlewareConfig{
				ProviderURL: "invalid",
				ClientID:    "1234567890",
			},
			expectedErr: errors.New("can't create new provider"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewAuthMiddleware(testCase.conf)
			if testCase.shouldFail {
				assert.Error(t, err)
				assert.ErrorContains(t, err, testCase.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthMiddlewareGuard(t *testing.T) {
	type testData struct {
		name                string
		conf                AuthMiddlewareConfig
		acceptedRoles       []string
		acceptedPermissions []string
		header              http.Header
		expectedStatusCode  int
	}

	provider := test.LaunchTestProvider()

	var testCases = [...]testData{
		{
			name: "Success case: only token check",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss": provider.URL,
						"exp": jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud": jwt.ClaimStrings{"123456"},
					}),
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Success case: role check",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
				AuthorizedRoles: ClaimCheckerConfig{
					TokenKey:  "role",
					ClaimType: "string",
				},
			},
			acceptedRoles: []string{
				"user",
				"manager",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss":  provider.URL,
						"exp":  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud":  jwt.ClaimStrings{"123456"},
						"role": "user",
					}),
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Success case: permission check",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
				RequiredPermissions: ClaimCheckerConfig{
					TokenKey:  "permission",
					ClaimType: "string",
				},
			},
			acceptedPermissions: []string{
				"read",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss":        provider.URL,
						"exp":        jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud":        jwt.ClaimStrings{"123456"},
						"permission": "read",
					}),
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Fail case: no token in header",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer",
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Fail case: expired token",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss": provider.URL,
						"exp": jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						"aud": jwt.ClaimStrings{"123456"},
					}),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Fail case: invalid role",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
				AuthorizedRoles: ClaimCheckerConfig{
					TokenKey:  "role",
					ClaimType: "string",
				},
			},
			acceptedRoles: []string{
				"manager",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss":  provider.URL,
						"exp":  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud":  jwt.ClaimStrings{"123456"},
						"role": "admin",
					}),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Fail case: invalid role key",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
				AuthorizedRoles: ClaimCheckerConfig{
					TokenKey:  "invalid",
					ClaimType: "string",
				},
			},
			acceptedRoles: []string{
				"manager",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss":  provider.URL,
						"exp":  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud":  jwt.ClaimStrings{"123456"},
						"role": "manager",
					}),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Fail case: invalid permission",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
				RequiredPermissions: ClaimCheckerConfig{
					TokenKey:  "permission",
					ClaimType: "string",
				},
			},
			acceptedPermissions: []string{
				"read",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss":        provider.URL,
						"exp":        jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud":        jwt.ClaimStrings{"123456"},
						"permission": "other",
					}),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Fail case: invalid permission key",
			conf: AuthMiddlewareConfig{
				ProviderURL: provider.URL,
				ClientID:    "123456",
				RequiredPermissions: ClaimCheckerConfig{
					TokenKey:  "invalid",
					ClaimType: "string",
				},
			},
			acceptedPermissions: []string{
				"read",
			},
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss":        provider.URL,
						"exp":        jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"aud":        jwt.ClaimStrings{"123456"},
						"permission": "read",
					}),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}
			c.Request.Header = testCase.header

			middleware, err := NewAuthMiddleware(testCase.conf)
			assert.NoError(t, err)

			middleware.Guard(testCase.acceptedRoles, testCase.acceptedPermissions)(c)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
		})
	}
}

func TestExtractToken(t *testing.T) {
	type testData struct {
		name          string
		header        http.Header
		expectedToken string
		expectedErr   error
	}

	var testCases = [...]testData{
		{
			name: "Success case",
			header: http.Header{
				"Authorization": []string{"Bearer token"},
			},
			expectedToken: "token",
			expectedErr:   nil,
		},
		{
			name:        "Fail case: no header",
			expectedErr: ErrNoAuthHeader,
		},
		{
			name: "Fail case: no auth header",
			header: http.Header{
				"Invalid": []string{"Bearer token"},
			},
			expectedErr: ErrNoAuthHeader,
		},
		{
			name: "Fail case: malformated auth header",
			header: http.Header{
				"Authorization": []string{"token"},
			},
			expectedErr: ErrMalformatedAuthHeader,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}
			c.Request.Header = testCase.header

			token, err := extractToken(c)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedToken, token)
		})
	}
}
