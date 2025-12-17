package main

import (
	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/middleware"
	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/routers"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	middleware.InitConfig(cfg)
	r := gin.Default()
	routers.Setup(r, cfg)
	r.Run(cfg.Addr)
}
