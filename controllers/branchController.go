package controllers

import (
	"errors"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllBranch(c *gin.Context) {
	var branches []models.Branch

	// Retrieve all branches from the database
	if err := database.DB.Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve branches"})
		return
	}

	// Return the list of branches
	c.JSON(http.StatusOK, gin.H{"data": branches})
}

func GetDetailsBranch(c *gin.Context) {
	// Get the branch ID from the URL parameter
	branchID := c.Param("BranchID")

	var branch models.Branch

	// Find the branch with the provided ID from the database
	if err := database.DB.Where("BranchID = ?", branchID).First(&branch).Error; err != nil {
		// If branch not found, return a 404 error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve branch details"})
		}
		return
	}

	// Return the branch details
	c.JSON(http.StatusOK, gin.H{"data": branch})
}

// AddBranch handles creating a new branch
func AddBranch(c *gin.Context) {
	// Xử lý upload file
	imageFile, err := c.FormFile("ImageURL")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ảnh không hợp lệ hoặc chưa được gửi lên"})
		return
	}

	filePath := "upload/" + imageFile.Filename
	if err := c.SaveUploadedFile(imageFile, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu file poster thất bại"})
		return
	}

	branch := models.Branch{
		BranchName:    c.Request.FormValue("BranchName"),
		Slug:          c.Request.FormValue("Slug"),
		Email:         c.Request.FormValue("Email"),
		Address:       c.Request.FormValue("Address"),
		PhoneNumber:   c.Request.FormValue("PhoneNumber"),
		ImageURL:      filePath,
		City:          c.Request.FormValue("City"),
		CreatedBy:     c.Request.FormValue("CreatedBy"),
		LastUpdatedBy: c.Request.FormValue("LastUpdatedBy"),
	}

	if err := database.DB.Create(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"data": branch})
}

// UpdateBranch handles updating an existing branch
func UpdateBranch(c *gin.Context) {
	var branch models.Branch

	// Get BranchID from URL parameters
	BranchID := c.Param("BranchID")

	// Find the branch by ID
	if err := database.DB.First(&branch, BranchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Bind incoming JSON to the Branch struct
	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update LastModified timestamp
	branch.LastUpdatedAt = time.Now()

	// Save changes to the database
	if err := database.DB.Save(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"data": branch})
}
