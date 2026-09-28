package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"node/global"
	"node/model/common"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				global.NODE_LOG.Error("panic recovered", zap.Any("error", err), zap.String("path", c.Request.URL.Path), zap.String("trace_id", GetTraceId(c)))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    common.Error_Internal_Server_Error,
					"message": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

const (
	traceIdKey    = "trace_id"
	traceIdHeader = "X-Trace-Id"
)

func TraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetHeader(traceIdHeader)
		if traceId == "" {
			traceId = generateTraceId()
		}
		c.Set(traceIdKey, traceId)
		c.Header(traceIdHeader, traceId)
		c.Next()
	}
}

func generateTraceId() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func GetTraceId(c *gin.Context) string {
	if v, ok := c.Get(traceIdKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		global.NODE_LOG.Info("request",
			zap.String(traceIdKey, GetTraceId(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("body_size", c.Writer.Size()),
		)
	}
}

func CORS(allowOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, o := range allowOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// if _, ok := allowed[origin]; ok {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Sec-WSS-Token, Connection, Upgrade, X-Trace-Id")
		// }

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
