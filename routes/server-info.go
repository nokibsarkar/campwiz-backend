package routes

import (
	"github.com/gin-gonic/gin"
)

func ServerInfoHeaderMiddleware(c *gin.Context) {
	c.Header("X-Backend-Served-By", serverInstanceId.String())
	c.Next()
}
