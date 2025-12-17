package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/services"
	"github.com/SpiritFoxo/control-system-microservices/shared/userroles"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(service *services.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// GetOrderById
// @Summary Gets order by ID
// @Description Gets order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param orderId path int true "Order ID"
// @Success 200 {object} services.OrderResponse "Order data"
// @Security BearerAuth
// @Router /orders/{orderId} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("orderId")
	var orderID uint
	if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	order, err := h.service.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetAllOrders
// @Summary Get list of orders
// @Description Retrieves a paginated list of orders with optional userId and status filters
// @Tags Orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of records per page" default(10)
// @Param userId query int false "Filter by user ID"
// @Param status query string false "Filter by order status"
// @Success 200 {object} map[string]interface{} "List of orders with pagination"
// @Security BearerAuth
// @Router /v1/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	rolesHeader := c.GetHeader("X-User-Roles")
	if rolesHeader == "" {
		if v, exists := c.Get("roles"); exists {
			if rs, ok := v.([]string); ok && len(rs) > 0 {
				rolesHeader = strings.Join(rs, ",")
			}
		}
	}

	userIdHeader := c.GetHeader("X-User-ID")
	var tokenUserID uint
	if userIdHeader != "" {
		if idInt, err := strconv.Atoi(userIdHeader); err == nil && idInt > 0 {
			tokenUserID = uint(idInt)
		}
	}
	roles := strings.Split(rolesHeader, ",")
	for i := range roles {
		roles[i] = strings.TrimSpace(roles[i])
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	userIDStr := c.Query("userId")
	statusFilter := c.Query("status")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit value"})
		return
	}

	isAllowedToViewAllOrders := false
	for _, role := range roles {
		if role != userroles.RoleEngineer {
			isAllowedToViewAllOrders = true
			break
		}
	}

	var userID uint
	if userIDStr != "" {
		userIDInt, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
			return
		}
		userID = uint(userIDInt)
	}

	if !isAllowedToViewAllOrders {
		if userID == 0 {
			userID = tokenUserID
		}
	}

	input := services.OrderListInput{
		Page:   page,
		Limit:  limit,
		UserID: userID,
		Status: statusFilter,
	}

	result, err := h.service.GetOrders(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": result.Orders,
			"pagination": gin.H{
				"page":       result.Page,
				"limit":      result.Limit,
				"total":      result.Total,
				"totalPages": result.TotalPages,
			},
		},
	})
}

// CreateOrder
// @Summary Create a new order
// @Description Creates a new order with provided data
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body services.CreateOrderInput true "Order creation data"
// @Success 201 {object} services.OrderResponse "Created order"
// @Security BearerAuth
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var input services.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.service.CreateOrder(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// UpdateOrderStatus
// @Summary Update an order
// @Description Updates an existing order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param orderId path int true "Order ID"
// @Param order body services.UpdateOrderInput true "Order update data"
// @Success 200 {object} services.OrderResponse "Updated order"
// @Security BearerAuth
// @Router /orders/{orderId} [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("orderId")
	var orderID uint
	if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.service.UpdateOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// CancelOrder
// @Summary Cancels an order
// @Description Cancels an existing order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param orderId path int true "Order ID"
// @Param order body services.UpdateOrderInput true "Order update data"
// @Success 200 {object} services.OrderResponse "Updated order"
// @Security BearerAuth
// @Router /orders/cancel/{orderId} [patch]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("orderId")
	var orderID uint
	if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	userIDStr := c.Request.Header.Get("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "missing X-User-ID"},
		})
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "broken X-User-ID"},
		})
		return
	}

	rolesStr := c.Request.Header.Get("X-User-Roles")
	if rolesStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "MISSING_HEADER", "message": "Отсутствует заголовок X-User-Roles"},
		})
		return
	}
	order, err := h.service.CancelOrder(orderID, uint(userID), rolesStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// DeleteOrder
// @Summary Delete an order
// @Description Deletes an order by ID
// @Tags Orders
// @Produce json
// @Param orderId path int true "Order ID"
// @Success 200 {object} map[string]string "Empty response"
// @Security BearerAuth
// @Router /orders/{orderId} [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id := c.Param("orderId")
	var orderID uint
	if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	if err := h.service.DeleteOrder(orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
