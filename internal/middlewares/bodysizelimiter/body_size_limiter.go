package bodysizelimiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Limit(maxBodyBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		c.Next()
	}
}
