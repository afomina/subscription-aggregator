package model

import (
	"time"
	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	CostRub     int        `json:"cost_rub" db:"cost_rub"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
}

type CreateSubscriptionRequest struct {
	ServiceName string     `json:"service_name" binding:"required"`
	CostRub     int        `json:"cost_rub" binding:"required,gt=0"`
	UserID      uuid.UUID  `json:"user_id" binding:"required"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string     `json:"service_name" binding:"required"`
	CostRub     int        `json:"cost_rub" binding:"required,gt=0"`
	UserID      uuid.UUID  `json:"user_id" binding:"required"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}