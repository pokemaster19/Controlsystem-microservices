package repositories

import "github.com/SpiritFoxo/control-system-microservices/service-users/internal/models"

type UserRepositoryInterface interface {
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User, updates map[string]interface{}) error
	GetUsers(page, limit int, emailFilter, roleFilter string) ([]models.User, int64, error)
}
