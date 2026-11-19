package middleware

import (
	"net/http"
	"node/global"
	"node/model/common"
	"strings"

	"github.com/gin-gonic/gin"
)

// DeshopWss 校验 Sec-WSS-Token。配置为空或与请求不匹配时一律拒绝。
func DeshopWss() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(global.NODE_CONFIG.Wss.SecWssToken)
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    common.Error_Service_Unavailable,
				"message": "wss token not configured",
			})
			return
		}

		provided := strings.TrimSpace(c.Request.Header.Get("Sec-WSS-Token"))
		if !secureEqual(provided, expected) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

// Wss 与 DeshopWss 相同，保留别名以免旧引用断裂。
func Wss() gin.HandlerFunc {
	return DeshopWss()
}
