package bodysizelimiter

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	type testData struct {
		name         string
		body         string
		maxBodyBytes int64
		shouldFail   bool
	}

	var testCases = [...]testData{
		{
			name:         "Success case: body too large",
			body:         "aaaaa",
			maxBodyBytes: 1,
			shouldFail:   true,
		},
		{
			name:         "Success case: body has valid size",
			body:         "aaaaa",
			maxBodyBytes: 1000000,
			shouldFail:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte(testCase.body)))
			c.Request = req
			Limit(testCase.maxBodyBytes)(c)

			_, err := io.ReadAll(c.Request.Body)
			if testCase.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
