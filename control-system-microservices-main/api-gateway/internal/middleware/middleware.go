package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var cfg *config.Config

func InitConfig(c *config.Config) {
	cfg = c
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "unauthorized", "message": "Missing or invalid Bearer token"},
			})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "unauthorized", "message": "Invalid token"},
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userIDStr := ""
			if idv, exists := claims["id"]; exists {
				userIDStr = fmt.Sprintf("%v", idv)
			}

			var roles []string
			switch raw := claims["roles"].(type) {
			case []interface{}:
				for _, r := range raw {
					if s, ok := r.(string); ok && s != "" {
						roles = append(roles, s)
					}
				}
			case []string:
				roles = raw
			case string:
				for _, s := range strings.Split(raw, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						roles = append(roles, s)
					}
				}
			default:
			}

			if userIDStr != "" {
				c.Request.Header.Set("X-User-ID", userIDStr)
			}
			if len(roles) > 0 {
				c.Request.Header.Set("X-User-Roles", strings.Join(roles, ","))
			}

			logger.Info("Authenticated request",
				zap.String("user_id", userIDStr),
				zap.Strings("roles", roles),
				zap.String("path", c.Request.URL.Path),
			)
		}
		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = "req-" + time.Now().Format("20060102150405")
		}
		c.Set("X-Request-ID", reqID)
		c.Header("X-Request-ID", reqID)
		c.Next()
	}
}

func RateLimiterManual() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  100,
	}
	store := memory.NewStore()
	lim := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	return func(c *gin.Context) {
		key := c.ClientIP()

		result, err := lim.Get(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}

		if result.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "rate_limit_exceeded",
					"message": "Too many requests",
				},
			})
			return
		}

		c.Next()
	}
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("req_id", c.GetString("X-Request-ID")),
		)
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, X-Request-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
