package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	// model "github.com/FloRichardAloeCorp/gateway/pkg/structs"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestForward(t *testing.T) {
	type requestParams struct {
		method string
		path   string
		query  string
		header map[string][]string
		body   io.Reader
	}

	type testData struct {
		name               string
		sourcePathPrefix   string
		handler            http.HandlerFunc
		req                requestParams
		expectedStatusCode int
	}

	var testCases = [...]testData{
		{
			name:             "Success case: method is forwarded",
			sourcePathPrefix: "/service",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}),
			req: requestParams{
				method: "POST",
				path:   "/service/test",
				body:   nil,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:             "Success case: path is forwarded",
			sourcePathPrefix: "/service",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/test" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}),
			req: requestParams{
				method: "GET",
				path:   "/service/test",
				body:   nil,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:             "Success case: query params are forwarded",
			sourcePathPrefix: "/service",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != "q=yes" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}),
			req: requestParams{
				method: "GET",
				path:   "/service/test",
				query:  "q=yes",
				body:   nil,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:             "Success case: headers are forwarded",
			sourcePathPrefix: "/service",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				testHeader := r.Header.Get("Test")
				if testHeader != "test" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}),
			req: requestParams{
				method: "GET",
				path:   "/service/test",
				header: map[string][]string{
					"Test": {"test"},
				},
				body: nil,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:             "Success case: body is forwarded",
			sourcePathPrefix: "/service",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				if string(body) != "test" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}),
			req: requestParams{
				method: "POST",
				path:   "/service/test",

				body: bytes.NewReader([]byte("test")),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:             "Fail case : invalid method",
			sourcePathPrefix: "/service",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusOK)
			}),
			req: requestParams{
				method: "GET",
				path:   "/service/test",
				body:   nil,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:             "Fail case: invalid path prefix",
			sourcePathPrefix: "/invalid",
			req: requestParams{
				method: "GET",
				path:   "/service/test",
				query:  "q=yes",
				body:   nil,
			},
			expectedStatusCode: http.StatusBadGateway,
		},
		{
			name:             "Fail case: malformated targetBaseURL",
			sourcePathPrefix: "/service",
			req: requestParams{
				method: "GET",
				path:   "/service/test",
				query:  "q=yes",
				body:   nil,
			},
			expectedStatusCode: http.StatusBadGateway,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(testCase.req.method, testCase.req.path+"?"+testCase.req.query, testCase.req.body)
			c.Request.Header = testCase.req.header

			server := httptest.NewServer(testCase.handler)
			defer server.Close()
			Init()
			Forward(testCase.sourcePathPrefix, server.URL)(c)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
		})
	}
}

func TestForwardInvalidTargetBaseURL(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://locahost:8080/service", nil)

	Init()
	Forward("/service", "unknown")(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	Forward("/service", "\t\n")(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestCopyResponseHeaders(t *testing.T) {
	response := &http.Response{Header: http.Header{"Test": []string{"value"}}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	copyResponseHeaders(response, c)

	assert.Equal(t, response.Header, w.Result().Header)
}
