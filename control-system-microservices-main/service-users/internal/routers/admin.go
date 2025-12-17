package routers

import (
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/handlers"
	"github.com/SpiritFoxo/control-system-microservices/shared/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(r *gin.RouterGroup, s *handlers.Server) {
	h := s.UserHandler

	r.POST("/users/register", middleware.RoleMiddleware(), h.RegisterUser)
	r.GET("/users/:userId", middleware.RoleMiddleware(), h.GetUserByID)
	r.PUT("/users/:userId", middleware.RoleMiddleware(), h.UpdateUser)
	r.GET("/users", middleware.RoleMiddleware(), h.GetUsers)
}
