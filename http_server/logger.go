package httpServer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/jxncyjq/stardust.mini/logs"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

var logger *zap.Logger

func initLogger() {
	if logger == nil {
		logger = logs.GetLogger("access middleware")
	}
}
func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Request Response 记录请求日志
func Request() echo.MiddlewareFunc {
	initLogger()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 取得request body
			requestBody := ""
			b, err := io.ReadAll(c.Request().Body)
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
				c.Request().Body = io.NopCloser(bytes.NewBuffer(b))
			}
			host := c.Request().Host
			uri := c.Request().RequestURI
			method := c.Request().Method
			agent := c.Request().UserAgent()

			// 取得 response body
			respBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, respBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer
			status := c.Response().Status
			ip := echo.ExtractIPFromXFFHeader()(c.Request())
			logger.Info("requests", zap.String("method", method),
				zap.String("uri", uri),
				zap.Int("status", status),
				zap.String("ip", ip),
				zap.String("agent", agent),
				zap.String("host", host),
				zap.String("body", requestBody))
			next(c)
			logger.Info("response",
				zap.String("response", respBody.String()),
			)
			return nil
		}
	}
}
