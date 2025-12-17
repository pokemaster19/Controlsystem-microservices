package main

import (
	"log"

	_ "github.com/SpiritFoxo/control-system-microservices/service-orders/docs"

	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/handlers"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/models"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/routers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func DbInit(cfg *config.Config) *gorm.DB {
	db, err := models.Setup(cfg)
	if err != nil {
		log.Println("Connection error")
	}
	return db
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	cfg := config.Load()
	db := DbInit(cfg)
	server := handlers.NewServer(db, cfg)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("api/v1")

	orders := api.Group("/orders")
	routers.SetupOrdersRoutes(orders, server)

	return r
}

// @title Orders Service API
// @version 1.0
// @description API for order management
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8081
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	r := SetupRouter()
	r.Run(":8081")
}
