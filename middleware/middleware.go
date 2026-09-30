package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"node/global"
	"node/model/common"
	"node/utils"
	"strconv"
	"strings"
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

const (
	maxLogBody = 4096
)

// 不记录 body 的路径(登录、改密等敏感接口)
var skipBodyPaths = map[string]struct{}{
	// "/api/login":          {},
	// "/api/change-password": {},
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		logBody := readBodyForLog(c)

		c.Next()

		latency := time.Since(start)

		fields := []zap.Field{
			zap.String(traceIdKey, GetTraceId(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("body", logBody),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.Int64("req_size", c.Request.ContentLength),
			zap.Int("resp_size", c.Writer.Size()),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		status := c.Writer.Status()
		switch {
		case status >= 500:
			global.NODE_LOG.Error("request", fields...)
			msg := fmt.Sprintf(
				"🚨 5xx Error\nTraceID: %s\n%s %s\nStatus: %d\nLatency: %s\nIP: %s\nErrors: %s\nBody: %s",
				GetTraceId(c),
				c.Request.Method,
				path,
				status,
				latency,
				c.ClientIP(),
				c.Errors.String(),
				logBody,
			)
			utils.InformToTelegram(msg)
		case status >= 400:
			global.NODE_LOG.Warn("request", fields...)
		default:
			global.NODE_LOG.Info("request", fields...)
		}
	}
}

func readBodyForLog(c *gin.Context) string {
	req := c.Request

	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return ""
	}
	if req.Body == nil || req.Body == http.NoBody {
		return ""
	}

	// 敏感路径不记录
	if _, ok := skipBodyPaths[req.URL.Path]; ok {
		return "[sensitive omitted]"
	}

	// 只记录文本类 body,文件上传/二进制直接跳过(先判断,再决定是否读取)
	ct := req.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") &&
		!strings.HasPrefix(ct, "application/x-www-form-urlencoded") &&
		!strings.HasPrefix(ct, "text/") {
		return "[" + ct + " omitted]"
	}

	// Content-Length 已知且过大:不读,避免占内存
	if req.ContentLength > maxLogBody {
		return "[body too large: " + strconv.FormatInt(req.ContentLength, 10) + " bytes]"
	}

	// 最多读 maxLogBody+1 字节,用来判断是否被截断
	buf, err := io.ReadAll(io.LimitReader(req.Body, maxLogBody+1))
	if err != nil {
		return "[read body error]"
	}
	// 把已读部分 + 剩余流 拼回去,保证后续 handler 能完整读取
	req.Body = struct {
		io.Reader
		io.Closer
	}{
		Reader: io.MultiReader(bytes.NewReader(buf), req.Body),
		Closer: req.Body,
	}

	if len(buf) > maxLogBody {
		return string(buf[:maxLogBody]) + "...[truncated]"
	}
	return string(buf)
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
