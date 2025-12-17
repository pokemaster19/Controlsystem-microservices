package routers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/handlers"
	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(r *gin.Engine, cfg *config.Config) {
	usersProxy := setupProxy(cfg.UsersServiceURL)
	r.POST("/api/v1/auth/login", middleware.RequestID(), middleware.Logging(), middleware.CORS(), middleware.RateLimiterManual(), usersProxy)
	r.Any("/api/v1/admin/*path", middleware.JWTAuth(), middleware.RequestID(), middleware.Logging(), middleware.CORS(), middleware.RateLimiterManual(), usersProxy)

	ordersProxy := setupProxy(cfg.OrdersServiceURL)
	r.Any("/api/v1/orders/*path", middleware.JWTAuth(), middleware.RequestID(), middleware.Logging(), middleware.CORS(), middleware.RateLimiterManual(), ordersProxy)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "API Gateway is running"})
	})
	r.GET("/swagger-docs/doc.json", handlers.SwaggerHandler(cfg))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger-docs/doc.json")))

}

func setupProxy(target string) gin.HandlerFunc {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.Director = func(req *http.Request) {
		originalPath := req.URL.Path

		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.URL.Path = url.Path + originalPath
		req.Host = url.Host
		req.Header.Set("X-Request-ID", req.Header.Get("X-Request-ID"))
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		writer.WriteHeader(http.StatusBadGateway)
		writer.Write([]byte("Service unavailable"))
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
