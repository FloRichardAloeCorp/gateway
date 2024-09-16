package headersizelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	// model "github.com/FloRichardAloeCorp/gateway/pkg/structs"
)

func TestLimit(t *testing.T) {
	type testData struct {
		name               string
		maxHeaderBytes     int
		header             http.Header
		expectedStatusCode int
	}

	var testCases = [...]testData{
		{
			name:           "Success case: header too large",
			maxHeaderBytes: 1,
			header: http.Header{
				"test": []string{"test"},
			},
			expectedStatusCode: http.StatusRequestHeaderFieldsTooLarge,
		},
		{
			name:           "Success case: header has valid size",
			maxHeaderBytes: 10000000,
			header: http.Header{
				"test": []string{"test"},
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/", nil)
			req.Header = testCase.header
			c.Request = req

			Limit(testCase.maxHeaderBytes)(c)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
		})
	}
}
