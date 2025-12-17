package models

import (
	"log"

	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/shared/userroles"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Setup(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Can not connect to the database:", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}

	var existingUser User
	if err := db.Where("email = ?", "admin@controlsystem.ru").First(&existingUser).Error; err == nil {
		log.Println("Admin user already exists, skipping creation")
	} else if err == gorm.ErrRecordNotFound {
		adminUser := &User{
			Name:     "Admin",
			Email:    "admin@controlsystem.ru",
			Roles:    []string{userroles.RoleSuperadmin},
			Password: cfg.SuperadminPassword,
		}

		if err := adminUser.HashPassword(); err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}
		if err := db.Create(adminUser).Error; err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		log.Println("Admin user created successfully with password " + cfg.SuperadminPassword)
	} else {
		log.Fatalf("Failed to check existing user: %v", err)
	}

	return db, nil
}
