package models

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
)

type OrderStatus int

const (
	StatusCreated OrderStatus = iota
	StatusAccepted
	StatusProcessed
	StatusClosed
	StatusCanceled
)

var statusNames = [...]string{
	"Created",
	"Accepted",
	"Processed",
	"Closed",
	"Canceled",
}

func (s OrderStatus) String() string {
	if int(s) < 0 || int(s) >= len(statusNames) {
		return "Unknown"
	}
	return statusNames[s]
}

type Order struct {
	gorm.Model
	UserId uint        `gorm:"not null"`
	Status OrderStatus `gorm:"type:order_status;not null;default:'Created'"`
	Cost   int         `gorm:"not null"`
	Items  []OrderItem
}

func (o *Order) NextStatus() error {
	if o.Status >= StatusClosed {
		return fmt.Errorf("Order is already closed")
	}
	o.Status++
	return nil
}

func (o *Order) Cancel() error {
	if o.Status == StatusClosed || o.Status == StatusCanceled {
		return fmt.Errorf("order cannot be canceled")
	}
	o.Status = StatusCanceled
	return nil
}

func (s *OrderStatus) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan OrderStatus from %T", value)
	}
	for i, v := range statusNames {
		if v == str {
			*s = OrderStatus(i)
			return nil
		}
	}
	return fmt.Errorf("invalid OrderStatus value: %s", str)
}

func (s OrderStatus) Value() (driver.Value, error) {
	if int(s) < 0 || int(s) >= len(statusNames) {
		return nil, fmt.Errorf("invalid OrderStatus index: %d", s)
	}
	return statusNames[s], nil
}

type OrderItem struct {
	gorm.Model
	OrderId  uint   `gorm:"not null"`
	Order    Order  `gorm:"foreignKey:OrderId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Quantity int    `gorm:"not null"`
	Name     string `gorm:"not null"`
}
