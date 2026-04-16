package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func SlogLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				slog.Error("http request error", "error", e)
			}
		} else {
			slog.Info("http request",
				"method", c.Request.Method,
				"path", path,
				"query", query,
				"status", status,
				"latency", latency,
				"ip", c.ClientIP(),
				"request_id", c.GetString("request_id"),
			)
		}
	}
}
