package middleware

import (
	"net/http"
	"node/global"

	"github.com/gin-gonic/gin"
)

func Wss() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Sec-WSS-Token")
		if token == "" || token != global.NODE_CONFIG.Wss.SecWssToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

func DeshopWss() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Sec-WSS-Token")
		if token == "" || token != global.NODE_CONFIG.Wss.SecWssToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
