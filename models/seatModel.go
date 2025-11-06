package models

type Seat struct {
	SeatID      int    `gorm:"primaryKey;autoIncrement;column:SeatID"`
	SeatNumber  int    `gorm:"not null;column:SeatNumber"`
	RowID       int    `gorm:"not null;column:RowID"`
	RowName     string `gorm:"not null;column:RowName"`
	Area        int    `gorm:"not null;column:Area"`
	Column      int    `gorm:"not null;column:Column"`
	Row         int    `gorm:"not null;column:Row"`
	Description string `gorm:"type:text;column:Description"`
	isOld       bool   `gorm:"column:IsOld;not null;default:false"`
}
