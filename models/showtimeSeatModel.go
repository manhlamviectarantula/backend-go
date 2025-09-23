package models

import "time"

type ShowtimeSeat struct {
	ShowtimeSeatID int       `gorm:"primaryKey;autoIncrement;column:ShowtimeSeatID"`
	ShowtimeID     int       `gorm:"not null;column:ShowtimeID"`
	SeatID         int       `gorm:"not null;column:SeatID"`
	RowName        string    `gorm:"not null;column:RowName"`
	TicketPrice    int       `gorm:"not null;column:TicketPrice"`
	Status         int8      `gorm:"type:tinyint(1);default:0;column:Status"`
	OrderID        int       `gorm:"column:OrderID;default:null"`
	LockedBy       int       `gorm:"column:LockedBy;default:null"`
	LockedAt       time.Time `gorm:"column:LockedAt;autoUpdateTime"`
}
