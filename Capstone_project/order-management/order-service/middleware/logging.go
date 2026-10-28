package middleware

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
	"time"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.New(slog.NewTextHandler(os.Stdout, nil)).Info("http_request", "request_id", c.GetString("request_id"), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration", time.Since(start).String())
	}
}
