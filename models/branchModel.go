package models

import "time"

type Branch struct {
	BranchID      int       `gorm:"primaryKey;autoIncrement;column:BranchID"`
	BranchName    string    `gorm:"size:100;not null;column:BranchName"`
	Slug          string    `gorm:"size:100;not null;column:Slug"`
	Email         string    `gorm:"size:100;unique;not null;column:Email"`
	Address       string    `gorm:"type:text;column:Address"`
	PhoneNumber   string    `gorm:"size:15;not null;column:PhoneNumber"`
	ImageURL      string    `gorm:"type:text;column:ImageURL"`
	City          string    `gorm:"size:255;not null;column:City"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:CreatedAt"`
	LastUpdatedAt time.Time `gorm:"autoUpdateTime;column:LastUpdatedAt"`
	CreatedBy     string    `gorm:"size:100;not null;column:CreatedBy"`
	LastUpdatedBy string    `gorm:"size:100;not null;column:LastUpdatedBy"`
}
