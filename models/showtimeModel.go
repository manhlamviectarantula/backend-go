package models

import "time"

type Showtime struct {
	ShowtimeID    int       `gorm:"primaryKey;autoIncrement;column:ShowtimeID"`
	TheaterID     int       `gorm:"not null;column:TheaterID"`
	MovieID       int       `gorm:"not null;column:MovieID"`
	ShowDate      string    `gorm:"not null;column:ShowDate"`
	StartTime     string    `gorm:"not null;column:StartTime"`
	EndTime       string    `gorm:"not null;column:EndTime"`
	Status        int       `gorm:"not null;column:Status;default:1"`
	IsOpenOrder   bool      `gorm:"not null;column:IsOpenOrder;default:false"`
	CancelReason  string    `gorm:"size:255;column:CancelReason"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:CreatedAt"`
	CreatedBy     string    `gorm:"size:100;not null;column:CreatedBy"`
	LastUpdatedAt time.Time `gorm:"autoUpdateTime;column:LastUpdatedAt"`
	LastUpdatedBy string    `gorm:"size:100;column:LastUpdatedBy"`

	Theater *Theater `gorm:"foreignKey:TheaterID;references:TheaterID"`
	Movie   *Movie   `gorm:"foreignKey:MovieID;references:MovieID"`
}
