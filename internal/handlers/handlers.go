package handlers

import (
	"github.com/gin-gonic/gin"
)

// NotFound returns custom 404 page
func NotFound(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 404,
		"msg":  "Not Found",
		"data": "",
	})
}
