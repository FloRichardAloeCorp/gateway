package ratelimiters

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/FloRichardAloeCorp/gateway/internal/test"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestSelectKeyRetriever(t *testing.T) {
	type testData struct {
		name        string
		limitBy     string
		expectedRes KeyRetriever
	}

	var testCases = [...]testData{
		{
			name:        "sub_claim retriever",
			limitBy:     "sub_claim",
			expectedRes: retrieveSubClaim,
		},
		{
			name:        "Default key retriever",
			limitBy:     "",
			expectedRes: defaultKeyRetriever,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			retriever := selectKeyRetriever(testCase.limitBy)
			assert.Equal(t,
				runtime.FuncForPC(reflect.ValueOf(testCase.expectedRes).Pointer()).Name(),
				runtime.FuncForPC(reflect.ValueOf(retriever).Pointer()).Name(),
			)
		})
	}
}

func TestRetrieveSubClaim(t *testing.T) {
	type testData struct {
		name        string
		header      http.Header
		expectedKey string
		shouldFail  bool
	}

	var testCases = [...]testData{
		{
			name: "Success case",
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss": "issuer",
						"sub": "id",
						"aud": jwt.ClaimStrings{"123456"},
						"exp": jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"nbf": jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						"iat": jwt.NewNumericDate(time.Now()),
						"jti": "id",
					}),
				},
			},
			expectedKey: "id",
		},
		{
			name: "Fail case: no sub claim in token",
			header: http.Header{
				"Authorization": []string{
					"Bearer " + test.NewToken(jwt.MapClaims{
						"iss": "issuer",
						"aud": jwt.ClaimStrings{"123456"},
						"exp": jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
						"nbf": jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						"iat": jwt.NewNumericDate(time.Now()),
						"jti": "id",
					}),
				},
			},
			shouldFail: true,
		},
		{
			name: "Fail case: no token",
			header: http.Header{
				"Authorization": []string{
					"Bearer ",
				},
			},
			shouldFail: true,
		},
		{
			name:       "Fail case: no auth header",
			header:     http.Header{},
			shouldFail: true,
		},
		{
			name: "Fail case: malformed token in auth header",
			header: http.Header{
				"Authorization": []string{
					"invalid",
				},
			},
			shouldFail: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}
			c.Request.Header = testCase.header

			key, err := retrieveSubClaim(c)
			if testCase.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedKey, key)
			}
		})
	}
}

func TestDefaultKeyRetriever(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	key, err := defaultKeyRetriever(c)
	assert.NoError(t, err)
	assert.Equal(t, "global", key)
}
