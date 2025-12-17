package handlers

import (
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/repositories"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/services"
	"gorm.io/gorm"
)

type Server struct {
	db           *gorm.DB
	cfg          *config.Config
	OrderService *services.OrderService

	OrderHandler *OrderHandler
}

func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository, cfg)
	orderHandler := NewOrderHandler(orderService)
	return &Server{
		db:           db,
		cfg:          cfg,
		OrderService: orderService,
		OrderHandler: orderHandler,
	}
}
