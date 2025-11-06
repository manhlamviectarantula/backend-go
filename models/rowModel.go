package models

type Row struct {
	RowID     int    `gorm:"primaryKey;autoIncrement;column:RowID"`
	TheaterID int    `gorm:"not null;column:TheaterID"`
	RowName   string `gorm:"size:100;not null;column:RowName"`
	isOld     bool   `gorm:"column:IsOld;not null;default:false"`
}
