package model

import (
	"time"
	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name" binding:"required"`
	CostRub     int        `json:"cost_rub" db:"cost_rub" binding:"required,gt=0"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id" binding:"required"`
	StartDate   time.Time  `json:"start_date" db:"start_date" binding:"required" time_format:"2006-01"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
}