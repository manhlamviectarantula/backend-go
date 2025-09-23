package models

import "time"

type Food struct {
	FoodID        int       `gorm:"column:FoodID;primaryKey;autoIncrement"`
	BranchID      int       `gorm:"column:BranchID;not null"`
	FoodName      string    `gorm:"column:FoodName;size:100;not null"`
	Image         string    `gorm:"column:Image;size:100;not null"`
	Description   string    `gorm:"column:Description;size:255;not null"`
	Price         int       `gorm:"column:Price;not null"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;autoCreateTime"`
	LastUpdatedAt time.Time `gorm:"column:LastUpdatedAt;autoUpdateTime"`
	CreatedBy     string    `gorm:"column:CreatedBy;size:100;not null"`
	LastUpdatedBy string    `gorm:"column:LastUpdatedBy;size:100;not null"`
}
