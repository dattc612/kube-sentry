package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func NamespaceGuardMiddleware(allowedNamespaces []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedNamespaces) == 0 {
			c.Next()
			return
		}

		targetNs := c.Param("namespace")
		if targetNs == "" {
			targetNs = c.Query("namespace")
		}

		if targetNs != "" && !slices.Contains(allowedNamespaces, targetNs) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access to namespace '" + targetNs + "' is restricted.",
			})
			return
		}

		c.Next()
	}
}
