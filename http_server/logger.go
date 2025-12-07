package httpServer

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust.mini/logs"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

var logger *zap.Logger

func initLogger() {
	if logger == nil {
		logger = logs.GetLogger("access middleware")
	}
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Request Response 记录请求日志
func Request() gin.HandlerFunc {
	initLogger()
	return func(c *gin.Context) {
		// 取得request body
		requestBody := ""
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			requestBody = "failed to request body"
		} else {
			var jsonData interface{}
			if err := json.Unmarshal(b, &jsonData); err != nil {
				// 不是有效的 JSON，原样输出
				requestBody = string(b)
			} else {
				// 格式化为标准 JSON 字符串
				formatted, _ := json.MarshalIndent(jsonData, "", "  ")
				requestBody = string(formatted)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		}
		host := c.Request.Host
		uri := c.Request.RequestURI
		method := c.Request.Method
		agent := c.Request.UserAgent()

		// 取得 response body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		status := c.Writer.Status()
		ip := c.ClientIP()
		logger.Info("requests", zap.String("method", method),
			zap.String("uri", uri),
			zap.Int("status", status),
			zap.String("ip", ip),
			zap.String("agent", agent),
			zap.String("host", host),
			zap.String("body", requestBody))
		c.Next()
		logger.Info("response",
			zap.String("response", blw.body.String()),
		)
	}
}
