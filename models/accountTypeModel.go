package models

type AccountType struct {
	AccountTypeID   int    `gorm:"primaryKey;autoIncrement"`
	AccountTypeName string `gorm:"size:255;not null"`
}
