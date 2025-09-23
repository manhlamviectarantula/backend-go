package models

import (
	"time"
)

type Movie struct {
	MovieID        int       `form:"MovieID" gorm:"primaryKey;autoIncrement;column:MovieID"`
	MovieName      string    `form:"MovieName" gorm:"size:255;not null;column:MovieName"`
	Slug           string    `form:"Slug" gorm:"size:100;not null;column:Slug"`
	AgeTag         string    `form:"AgeTag" gorm:"size:15;not null;column:AgeTag"`
	Duration       int       `form:"Duration" gorm:"not null;column:Duration"`
	ReleaseDate    string    `form:"ReleaseDate" gorm:"size:100;not null;column:ReleaseDate"`
	LastScreenDate string    `form:"LastScreenDate" gorm:"size:100;not null;column:LastScreenDate"`
	Poster         string    `form:"Poster" binding:"required" gorm:"type:text;column:Poster"`
	Trailer        string    `form:"Trailer" gorm:"type:text;column:Trailer"`
	Rating         float64   `form:"Rating" gorm:"type:decimal(10,2);not null;column:Rating"`
	Description    string    `form:"Description" gorm:"type:text;not null;column:Description"`
	Status         int       `form:"Status" gorm:"not null;column:Status"`
	CreatedAt      time.Time `form:"CreatedAt" gorm:"autoCreateTime;column:CreatedAt"`
	LastUpdatedAt  time.Time `form:"LastUpdatedAt" gorm:"autoUpdateTime;column:LastUpdatedAt"`
	CreatedBy      string    `form:"CreatedBy" gorm:"size:100;not null;column:CreatedBy"`
	LastUpdatedBy  string    `form:"LastUpdatedBy" gorm:"size:100;not null;column:LastUpdatedBy"`
}
