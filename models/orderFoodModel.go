package models

type OrderFood struct {
	OrderFoodID int `gorm:"column:OrderFoodID;primaryKey;autoIncrement"`
	OrderID     int `gorm:"column:OrderID;not null"`
	FoodID      int `gorm:"column:FoodID;not null"`
	Quantity    int `gorm:"column:Quantity;not null"`
	TotalPrice  int `gorm:"column:TotalPrice;not null"`
}
