package middleware

import (
	"net/http"
	"strings"

	userroles "github.com/SpiritFoxo/control-system-microservices/shared/userroles"
	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesHeader := c.GetHeader("X-User-Roles")
		if rolesHeader == "" {
			if v, exists := c.Get("roles"); exists {
				if rs, ok := v.([]string); ok && len(rs) > 0 {
					rolesHeader = strings.Join(rs, ",")
				}
			}
		}

		if rolesHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "unauthorized", "message": "Missing roles header"},
			})
			c.Abort()
			return
		}

		raw := strings.Split(rolesHeader, ",")
		var roles []string
		for _, r := range raw {
			if s := strings.TrimSpace(r); s != "" {
				roles = append(roles, s)
			}
		}

		for _, r := range roles {
			if r == userroles.RoleAdmin || r == userroles.RoleSuperadmin {
				c.Next()
				return
			}
		}

		for _, userRole := range roles {
			for _, allowed := range allowedRoles {
				if userRole == allowed {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   gin.H{"code": "forbidden", "message": "forbidden"},
		})
		c.Abort()
	}
}
