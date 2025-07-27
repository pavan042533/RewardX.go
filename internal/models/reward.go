package models

import (
	"time"
)

type Reward struct {
	ID                        uint      `gorm:"primaryKey" json:"id"`
	Name                      string    `gorm:"unique"    json:"name"`
	Category                  string    `json:"category"`
	Cost                      int       `json:"cost"`
	Stock                     int       `json:"stock"`
	CreatedByID               uint      `json:"created_by_id"`
	Discount                  float64   `json:"discount"`
	CampaignName              string    `json:"campaign_name"`
	Description               string    `json:"description"`
	StartDate                 time.Time `json:"start_date"`
	EndDate                   time.Time `json:"end_date"`
	AutoExpireAfterRedemption bool      `json:"auto_expire_after_redemption"`
}

