package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/pkg/langfuse"
)

func Tracing(client *langfuse.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		latencyMs := time.Since(start).Milliseconds()
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		traceName := fmt.Sprintf("http-%s-%s", method, path)

		userIDStr := ""
		if uid := UserIDFromContext(c); uid != 0 {
			userIDStr = fmt.Sprintf("%d", uid)
		}

		reqID, _ := c.Get("request_id")
		reqIDStr, _ := reqID.(string)

		traceID := client.CreateTrace(&langfuse.TraceBody{
			Name:   traceName,
			UserID: userIDStr,
			Tags:   []string{"http", fmt.Sprintf("status:%d", statusCode), method},
			Metadata: map[string]interface{}{
				"request_id": reqIDStr,
				"latency_ms": latencyMs,
				"client_ip":  c.ClientIP(),
			},
		})

		client.CreateSpan(&langfuse.ObservationBody{
			TraceID: traceID,
			Name:    traceName,
			Input: map[string]interface{}{
				"method":      method,
				"path":        path,
				"query":       c.Request.URL.RawQuery,
				"request_id":  reqIDStr,
				"user_agent":  c.GetHeader("User-Agent"),
			},
			Output: map[string]interface{}{
				"status":    statusCode,
				"latency_ms": latencyMs,
			},
			Metadata: map[string]interface{}{
				"latency_ms": latencyMs,
				"request_id": reqIDStr,
			},
		})

		client.FlushAsync()
	}
}
