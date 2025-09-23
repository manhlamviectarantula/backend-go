package controllers

import (
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddSeats(c *gin.Context) {
	var seats []models.Seat

	// Bind incoming JSON to a slice of Seat structs
	if err := c.ShouldBindJSON(&seats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Disable foreign key checks
	if err := database.DisableForeignKeyChecks(database.DB, c); err != nil {
		return
	}

	// Save the new seats to the database
	if err := database.DB.Create(&seats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create seats"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"data": seats})
}
