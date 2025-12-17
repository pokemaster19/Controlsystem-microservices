package services

import (
	"testing"

	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/models"
	"github.com/SpiritFoxo/control-system-microservices/service-orders/internal/repositories/mocks"
	"github.com/SpiritFoxo/control-system-microservices/shared/userroles"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func newTestOrder(id uint, userID uint, status models.OrderStatus, cost int, items ...models.OrderItem) *models.Order {
	return &models.Order{
		Model:  gorm.Model{ID: id},
		UserId: userID,
		Status: status,
		Cost:   cost,
		Items:  items,
	}
}

func newTestOrderItem(id uint, name string, quantity int) models.OrderItem {
	return models.OrderItem{
		Model:    gorm.Model{ID: id},
		Name:     name,
		Quantity: quantity,
	}
}

func setupOrderTest(t *testing.T) (*OrderService, *mocks.MockOrderRepositoryInterface, func()) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryInterface(ctrl)
	cfg := &config.Config{}
	service := NewOrderService(mockRepo, cfg)
	return service, mockRepo, ctrl.Finish
}

func TestOrderService_GetOrderByID(t *testing.T) {
	service, mockRepo, finish := setupOrderTest(t)
	defer finish()
	item := newTestOrderItem(1, "Laptop", 2)
	order := newTestOrder(1, 100, models.StatusCreated, 2000, item)
	tests := []struct {
		name        string
		id          uint
		setupMock   func()
		expected    *OrderResponse
		expectedErr string
	}{
		{
			name: "успешно",
			id:   1,
			setupMock: func() {
				mockRepo.EXPECT().GetOrderByID(uint(1)).Return(order, nil)
			},
			expected: &OrderResponse{
				ID:     1,
				UserID: 100,
				Status: models.StatusCreated,
				Cost:   2000,
				OrderItems: []OrderItemResponse{
					{ID: 1, Name: "Laptop", Quantity: 2},
				},
			},
		},
		{
			name: "не найден",
			id:   999,
			setupMock: func() {
				mockRepo.EXPECT().GetOrderByID(uint(999)).Return((*models.Order)(nil), assert.AnError)
			},
			expectedErr: "assert.AnError",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := service.GetOrderByID(tt.id)
			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	service, mockRepo, finish := setupOrderTest(t)
	defer finish()
	input := &CreateOrderInput{
		UserID: 100,
		Status: models.StatusCreated,
		Cost:   2000,
		OrderItems: []OrderItemInput{
			{Name: "Laptop", Quantity: 2},
		},
	}
	tests := []struct {
		name        string
		input       *CreateOrderInput
		setupMock   func()
		expected    *OrderResponse
		expectedErr string
	}{
		{
			name:  "успешно",
			input: input,
			setupMock: func() {
				mockRepo.EXPECT().CreateOrder(gomock.Any()).DoAndReturn(func(o *models.Order) error {
					o.ID = 1
					o.Items[0].ID = 1
					return nil
				})
			},
			expected: &OrderResponse{
				ID:     1,
				UserID: 100,
				Status: models.StatusCreated,
				Cost:   2000,
				OrderItems: []OrderItemResponse{
					{ID: 1, Name: "Laptop", Quantity: 2},
				},
			},
		},
		{
			name: "пустые items",
			input: &CreateOrderInput{
				UserID:     100,
				Status:     models.StatusCreated,
				Cost:       2000,
				OrderItems: []OrderItemInput{},
			},
			setupMock:   func() {},
			expectedErr: "binding",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := service.CreateOrder(tt.input)
			if tt.expectedErr != "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
func TestOrderService_UpdateOrder(t *testing.T) {
	service, mockRepo, finish := setupOrderTest(t)
	defer finish()
	tests := []struct {
		name          string
		id            uint
		initialStatus models.OrderStatus
		setupMock     func(initialOrder *models.Order)
		expected      models.OrderStatus
		expectedErr   string
	}{
		{
			name:          "Created → Accepted",
			id:            1,
			initialStatus: models.StatusCreated,
			setupMock: func(initialOrder *models.Order) {
				updatedOrder := *initialOrder
				updatedOrder.Status = models.StatusAccepted
				gomock.InOrder(
					mockRepo.EXPECT().GetOrderByID(uint(1)).Return(initialOrder, nil),
					mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(nil),
					mockRepo.EXPECT().GetOrderByID(uint(1)).Return(&updatedOrder, nil),
				)
			},
			expected: models.StatusAccepted,
		},
		{
			name:          "Accepted → Processed",
			id:            2,
			initialStatus: models.StatusAccepted,
			setupMock: func(initialOrder *models.Order) {
				gomock.InOrder(
					mockRepo.EXPECT().GetOrderByID(uint(2)).Return(initialOrder, nil),
					mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(nil),
					mockRepo.EXPECT().GetOrderByID(uint(2)).Return(initialOrder, nil),
				)
			},
			expected: models.StatusProcessed,
		},
		{
			name:          "Processed → Closed",
			id:            3,
			initialStatus: models.StatusProcessed,
			setupMock: func(initialOrder *models.Order) {
				gomock.InOrder(
					mockRepo.EXPECT().GetOrderByID(uint(3)).Return(initialOrder, nil),
					mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(nil),
					mockRepo.EXPECT().GetOrderByID(uint(3)).Return(initialOrder, nil),
				)
			},
			expected: models.StatusClosed,
		},
		{
			name:          "Closed → ошибка",
			id:            4,
			initialStatus: models.StatusClosed,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(4)).Return(initialOrder, nil)
			},
			expectedErr: "order is already closed",
		},
		{
			name:          "Canceled → ошибка",
			id:            5,
			initialStatus: models.StatusCanceled,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(5)).Return(initialOrder, nil)
			},
			expectedErr: "order is already closed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialOrder := newTestOrder(tt.id, 100, tt.initialStatus, 2000, newTestOrderItem(1, "Laptop", 2))
			tt.setupMock(initialOrder)
			resp, err := service.UpdateOrder(tt.id)
			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expected, resp.Status)
				assert.Len(t, resp.OrderItems, 1)
			}
		})
	}
}

func TestOrderService_CancelOrder(t *testing.T) {
	service, mockRepo, finish := setupOrderTest(t)
	defer finish()
	tests := []struct {
		name          string
		id            uint
		userID        uint
		rolesStr      string
		initialStatus models.OrderStatus
		setupMock     func(initialOrder *models.Order)
		expected      models.OrderStatus
		expectedErr   string
	}{
		{
			name:          "инженер отменяет свой заказ",
			id:            1,
			userID:        100,
			rolesStr:      userroles.RoleEngineer,
			initialStatus: models.StatusCreated,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(1)).Return(initialOrder, nil)
				mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(nil)
			},
			expected: models.StatusCanceled,
		},
		{
			name:          "инженер пытается отменить чужой заказ",
			id:            1,
			userID:        999,
			rolesStr:      userroles.RoleEngineer,
			initialStatus: models.StatusCreated,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(1)).Return(initialOrder, nil)
			},
			expectedErr: "access forbidden",
		},
		{
			name:          "менеджер отменяет любой заказ",
			id:            1,
			userID:        999,
			rolesStr:      userroles.RoleManager,
			initialStatus: models.StatusCreated,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(1)).Return(initialOrder, nil)
				mockRepo.EXPECT().UpdateOrder(gomock.Any()).Return(nil)
			},
			expected: models.StatusCanceled,
		},
		{
			name:          "менеджер не может отменить закрытый заказ",
			id:            3,
			userID:        999,
			rolesStr:      userroles.RoleManager,
			initialStatus: models.StatusClosed,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(3)).Return(initialOrder, nil)
			},
			expectedErr: "order cannot be canceled",
		},
		{
			name:          "нельзя отменить уже отменённый заказ",
			id:            4,
			userID:        100,
			rolesStr:      userroles.RoleEngineer,
			initialStatus: models.StatusCanceled,
			setupMock: func(initialOrder *models.Order) {
				mockRepo.EXPECT().GetOrderByID(uint(4)).Return(initialOrder, nil)
			},
			expectedErr: "order cannot be canceled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialOrder := newTestOrder(tt.id, 100, tt.initialStatus, 2000)
			tt.setupMock(initialOrder)
			resp, err := service.CancelOrder(tt.id, tt.userID, tt.rolesStr)
			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expected, resp.Status)
			}
		})
	}
}

func TestOrderService_DeleteOrder(t *testing.T) {
	service, mockRepo, finish := setupOrderTest(t)
	defer finish()
	order := newTestOrder(1, 100, models.StatusCreated, 2000)
	tests := []struct {
		name        string
		id          uint
		setupMock   func()
		expectedErr string
	}{
		{
			name: "успешно",
			id:   1,
			setupMock: func() {
				mockRepo.EXPECT().GetOrderByID(uint(1)).Return(order, nil)
				mockRepo.EXPECT().DeleteOrder(order).Return(nil)
			},
		},
		{
			name: "не найден",
			id:   999,
			setupMock: func() {
				mockRepo.EXPECT().GetOrderByID(uint(999)).Return((*models.Order)(nil), assert.AnError)
			},
			expectedErr: "assert.AnError",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := service.DeleteOrder(tt.id)
			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderService_GetOrders(t *testing.T) {
	service, mockRepo, finish := setupOrderTest(t)
	defer finish()
	order1 := newTestOrder(1, 100, models.StatusCreated, 2000, newTestOrderItem(1, "Laptop", 2))
	order2 := newTestOrder(2, 101, models.StatusAccepted, 3000, newTestOrderItem(2, "Mouse", 5))
	orders := []models.Order{*order1, *order2}
	tests := []struct {
		name        string
		input       OrderListInput
		setupMock   func()
		expectedLen int
		expectedErr string
	}{
		{
			name:  "успешно",
			input: OrderListInput{Page: 1, Limit: 10},
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(1, 10, uint(0), "").Return(orders, int64(2), nil)
			},
			expectedLen: 2,
		},
		{
			name:  "фильтр по пользователю",
			input: OrderListInput{Page: 1, Limit: 10, UserID: 100},
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(1, 10, uint(100), "").Return([]models.Order{*order1}, int64(1), nil)
			},
			expectedLen: 1,
		},
		{
			name:  "фильтр по статусу",
			input: OrderListInput{Page: 1, Limit: 10, Status: "Accepted"},
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(1, 10, uint(0), "Accepted").Return([]models.Order{*order2}, int64(1), nil)
			},
			expectedLen: 1,
		},
		{
			name:        "невалидная страница",
			input:       OrderListInput{Page: 0, Limit: 10},
			setupMock:   func() {},
			expectedErr: "invalid page number",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := service.GetOrders(tt.input)
			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result.Orders, tt.expectedLen)
				assert.Equal(t, int64(tt.expectedLen), result.Total)
			}
		})
	}
}
