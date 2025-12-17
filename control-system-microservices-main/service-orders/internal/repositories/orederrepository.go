package repositories

import (
	"errors"

	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) GetOrderByID(id uint) (*models.Order, error) {
	var order models.Order
	result := r.db.Preload("Items").First(&order, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, result.Error
	}
	return &order, nil
}

func (r *OrderRepository) UpdateOrder(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) DeleteOrder(order *models.Order) error {
	return r.db.Delete(order).Error
}

func (r *OrderRepository) GetOrders(page, limit int, userID uint, status string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Offset((page - 1) * limit).
		Limit(limit).
		Preload("Items").
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
