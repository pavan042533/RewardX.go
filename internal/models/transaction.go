package models
import (
	"time"
)

type Transaction struct{
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	RewardID    uint      `json:"reward_id"`
	Status      string    `json:"status"`
	CouponCode string    `json:"coupon_code"`
	PointsUsed  int       `json:"points_used"`
	CreatedAt   time.Time `json:"created_at"` 
}