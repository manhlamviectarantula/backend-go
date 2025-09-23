package controllers

import (
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddOrderFood(c *gin.Context) {
	var orderFood models.OrderFood

	orderIDStr := c.Param("OrderID")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrderID is required"})
		return
	}

	// Chuyển string -> int
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OrderID format"})
		return
	}

	// Parse JSON
	if err := c.ShouldBindJSON(&orderFood); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gán OrderID đúng kiểu int
	orderFood.OrderID = orderID

	// Lưu vào DB
	if err := database.DB.Create(&orderFood).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order food"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Order food created successfully",
		"orderFood": orderFood,
	})
}
