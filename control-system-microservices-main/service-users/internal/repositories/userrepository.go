package repositories

import (
	"fmt"
	"strings"

	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", strings.ToLower(email)).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(user *models.User, updates map[string]interface{}) error {
	return r.db.Model(user).Updates(updates).Error
}

func (r *UserRepository) DeleteUser(user *models.User) error {
	return r.db.Delete(user).Error
}

func (r *UserRepository) GetUsers(page, limit int, emailFilter, roleFilter string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	if emailFilter != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(emailFilter)+"%")
	}

	if roleFilter != "" {
		query = query.Where("? = ANY(roles)", roleFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %v", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users: %v", err)
	}

	return users, total, nil
}
