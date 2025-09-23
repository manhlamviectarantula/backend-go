package controllers

import (
	"errors"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DeleteRow(c *gin.Context) {
	// Get the ID of the row to delete from the URL parameter
	rowID := c.Param("id")

	// Find the row by ID
	var row models.Row
	if err := database.DB.First(&row, rowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return not found error if row doesn't exist
			c.JSON(http.StatusNotFound, gin.H{"error": "Row not found"})
			return
		}
		// Return internal server error if there's another issue
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find row"})
		return
	}

	// Delete the row from the database
	if err := database.DB.Delete(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete row"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Row deleted successfully"})
}

func AddRows(c *gin.Context) {
	var rows []models.Row

	// Bind incoming JSON to the slice of Row structs
	if err := c.ShouldBindJSON(&rows); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Disable foreign key checks if needed
	if err := database.DisableForeignKeyChecks(database.DB, c); err != nil {
		return
	}

	// Save all rows to the database in a single transaction
	if err := database.DB.Create(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rows"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"data": rows})
}
