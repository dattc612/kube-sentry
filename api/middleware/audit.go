package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func AuditMiddleware(auditLogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		user := c.GetString("user_id")
		if user == "" {
			user = c.GetString("username")
		}
		if user == "" {
			user = "anonymous"
		}
		action, exists := c.Get("action")
		if !exists {
			action = "unknown"
		}

		// var rolesFormatted string
		// if rolesRaw, exists := c.Get("roles"); exists {
		// 	switch v := rolesRaw.(type) {
		// 	case []string:
		// 		rolesFormatted = strings.Join(v, ",")
		// 	case string:
		// 		rolesFormatted = v
		// 	}
		// }

		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		auditLogger.Info("api_access_event",
			slog.String("timestamp", start.Format(time.RFC3339)),
			slog.String("actor", user),
			// slog.String("roles", rolesFormatted),
			slog.String("method", c.Request.Method),
			slog.String("host", c.Request.Host),
			slog.String("path", path),
			slog.String("action", action.(string)),
			slog.String("matched_route", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.String("latency", time.Since(start).String()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
