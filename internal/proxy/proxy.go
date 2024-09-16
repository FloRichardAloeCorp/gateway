package proxy

import (
	"io"
	"net/http"
	"strings"

	"github.com/Aloe-Corporation/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	client *http.Client
	log    = logs.Get()
)

func Init() {
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func Forward(sourcePathPrefix, targetBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetURL, err := buildTargetURL(c, sourcePathPrefix, targetBaseURL)
		if err != nil {
			log.Error("Forward failure", zap.Error(err))
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}

		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			log.Error("Forward failure", zap.Error(err))
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}

		req.Header = c.Request.Header

		res, err := client.Do(req)
		if err != nil {
			log.Error("Forward failure", zap.Error(err))
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}
		defer res.Body.Close()

		_, err = io.Copy(c.Writer, res.Body)
		if err != nil {
			log.Error("Forward failure", zap.Error(err))
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}

		c.Status(res.StatusCode)
		copyResponseHeaders(res, c)
		c.Writer.WriteHeaderNow()
	}
}

func buildTargetURL(c *gin.Context, sourcePathPrefix, targetBaseURL string) (string, error) {
	if !strings.HasPrefix(c.Request.URL.Path, sourcePathPrefix) {
		return "", ErrUnknownPathPrefix
	}

	return targetBaseURL + strings.TrimPrefix(c.Request.URL.Path, sourcePathPrefix), nil
}

func copyResponseHeaders(res *http.Response, c *gin.Context) {
	for key, values := range res.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
}
