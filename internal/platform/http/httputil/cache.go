package httputil

import (
	"github.com/gin-gonic/gin"
)

func SetCacheHeader(c *gin.Context, status string) {
	if status != "" {
		c.Header("X-Cache", status)
	}
}
