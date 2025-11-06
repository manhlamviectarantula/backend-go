package models

import "time"

type Theater struct {
	TheaterID     int       `gorm:"primaryKey;autoIncrement;column:TheaterID"`
	BranchID      int       `gorm:"not null;column:BranchID"`
	TheaterName   string    `gorm:"size:100;not null;column:TheaterName"`
	Slug          string    `gorm:"size:100;not null;column:Slug"`
	TheaterType   string    `gorm:"size:10;not null;column:TheaterType"`
	MaxRow        int       `gorm:"not null;column:MaxRow"`
	MaxColumn     int       `gorm:"not null;column:MaxColumn"`
	SeatsPrice    int       `gorm:"not null;default:50000;column:SeatsPrice"`
	Status        bool      `gorm:"not null;default:true;column:Status"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:CreatedAt"`
	LastUpdatedAt time.Time `gorm:"autoUpdateTime;column:LastUpdatedAt"`
	CreatedBy     string    `gorm:"size:100;not null;column:CreatedBy"`
	LastUpdatedBy string    `gorm:"size:100;not null;column:LastUpdatedBy"`
}

// type Theater struct {
// 	TheaterID      int       `json:"theater_id" gorm:"primaryKey"`          // Theater ID (PK)
// 	BranchID       int       `json:"branch_id" gorm:"foreignKey:branch_id"` // Branch ID
// 	TheaterName    string    `json:"theater_name"`                          // Theater Name
// 	Slug           string    `json:"slug"`                                  // Identifier
// 	TheaterType    string    `json:"theater_type"`                          // Theater Type
// 	TotalSeats     int       `json:"total_seats"`                           // Total Seats
// 	CreatedAt      time.Time `json:"created_at"`                            // Creation Date
// 	CreatedBy      string    `json:"created_by"`                            // Created By
// 	LastModified   time.Time `json:"last_modified"`                         // Last Modified Date
// 	LastModifiedBy string    `json:"last_modified_by"`                      // Last Modified By

// 	Branch Branch `gorm:"foreignKey:BranchID;references:BranchID"`
// }
