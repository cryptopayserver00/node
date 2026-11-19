package middleware

import (
	"crypto/subtle"
	"net/http"
	"node/global"
	"node/model/common"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	headerAPIToken   = "X-API-Token"
	headerAuth       = "Authorization"
	bearerPrefix     = "Bearer "
)

// APIAuth 校验服务间共享密钥。接受 Authorization: Bearer <token> 或 X-API-Token。
// api-token 未配置时 fail-closed，返回 503。
func APIAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(global.NODE_CONFIG.Security.ApiToken)
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    common.Error_Service_Unavailable,
				"message": "api token not configured",
			})
			return
		}

		provided := extractAPIToken(c)
		if !secureEqual(provided, expected) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    common.Error_Unauthorized,
				"message": "unauthorized",
			})
			return
		}

		c.Next()
	}
}

func extractAPIToken(c *gin.Context) string {
	if token := strings.TrimSpace(c.GetHeader(headerAPIToken)); token != "" {
		return token
	}
	auth := c.GetHeader(headerAuth)
	if strings.HasPrefix(auth, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(auth, bearerPrefix))
	}
	return ""
}

func secureEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
