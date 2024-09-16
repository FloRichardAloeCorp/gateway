package headersizelimiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Limit(maxHeaderBytes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerSize := 0

		for key, val := range c.Request.Header {
			headerSize += len(key) + len(val)
			if headerSize > maxHeaderBytes {
				c.AbortWithStatusJSON(http.StatusRequestHeaderFieldsTooLarge, "headers too large")
				return
			}
		}

		c.Next()
	}
}
