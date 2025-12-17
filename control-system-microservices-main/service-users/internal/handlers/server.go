package handlers

import (
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/repositories"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/services"
	"gorm.io/gorm"
)

type Server struct {
	db          *gorm.DB
	cfg         *config.Config
	UserService *services.UserService

	UserHandler *UserHandler
}

func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository, cfg)
	userHandler := NewUserHandler(userService)
	return &Server{
		db:          db,
		cfg:         cfg,
		UserService: userService,
		UserHandler: userHandler,
	}
}
