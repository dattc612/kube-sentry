package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func RequireAction(action string, permissionsMap map[string][]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("action", action)
		allowedRoles, exists := permissionsMap[action]
		if !exists {
			c.Next()
			return
		}

		userRolesRaw, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: No roles found in context",
			})
			return
		}

		userRoles, ok := userRolesRaw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid role context format",
			})
			return
		}

		for _, role := range userRoles {
			if slices.Contains(allowedRoles, role) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Access denied: Insufficient permissions for action " + action,
		})
	}
}
