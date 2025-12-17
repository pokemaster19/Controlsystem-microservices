package repositories

import "github.com/SpiritFoxo/control-system-microservices/service-orders/internal/models"

type OrderRepositoryInterface interface {
	GetOrderByID(id uint) (*models.Order, error)
	CreateOrder(order *models.Order) error
	UpdateOrder(order *models.Order) error
	DeleteOrder(order *models.Order) error
	GetOrders(page, limit int, userID uint, status string) ([]models.Order, int64, error)
}
