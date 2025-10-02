package controllers

import (
	"errors"
	"log"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"movie-ticket-booking/services"
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

func AddBranch(c *gin.Context) {
	// Lấy file ảnh từ form
	imageFile, err := c.FormFile("ImageURL")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ảnh là bắt buộc"})
		return
	}

	imageURL, err := services.UploadToCloudinary(imageFile, "branches")
	if err != nil {
		log.Println("Upload Cloudinary error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload ảnh thất bại"})
		return
	}

	if imageURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload ảnh thất bại, URL rỗng"})
		return
	}

	// Tạo branch mới
	branch := models.Branch{
		BranchName:    c.PostForm("BranchName"),
		Slug:          c.PostForm("Slug"),
		Email:         c.PostForm("Email"),
		Address:       c.PostForm("Address"),
		PhoneNumber:   c.PostForm("PhoneNumber"),
		ImageURL:      imageURL,
		City:          c.PostForm("City"),
		CreatedBy:     c.PostForm("CreatedBy"),
		LastUpdatedBy: c.PostForm("LastUpdatedBy"),
	}

	// Lưu vào database
	if err := database.DB.Create(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tạo branch thất bại"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Branch added successfully", "data": branch})
}

func UpdateBranch(c *gin.Context) {
	var branch models.Branch

	// Lấy BranchID từ URL
	BranchID := c.Param("BranchID")

	// Tìm branch theo ID
	if err := database.DB.First(&branch, BranchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Bind dữ liệu từ form (không bind trực tiếp vào branch cũ để tránh ghi đè ID)
	branchName := c.PostForm("BranchName")
	slug := c.PostForm("Slug")
	email := c.PostForm("Email")
	address := c.PostForm("Address")
	phoneNumber := c.PostForm("PhoneNumber")
	city := c.PostForm("City")
	lastUpdatedBy := c.PostForm("LastUpdatedBy")

	// Cập nhật các trường (nếu có)
	if branchName != "" {
		branch.BranchName = branchName
	}
	if slug != "" {
		branch.Slug = slug
	}
	if email != "" {
		branch.Email = email
	}
	if address != "" {
		branch.Address = address
	}
	if phoneNumber != "" {
		branch.PhoneNumber = phoneNumber
	}
	if city != "" {
		branch.City = city
	}
	if lastUpdatedBy != "" {
		branch.LastUpdatedBy = lastUpdatedBy
	}

	// Kiểm tra có file ảnh mới không
	imageFile, err := c.FormFile("ImageURL")
	if err == nil && imageFile != nil {
		imageURL, err := services.UploadToCloudinary(imageFile, "branches")
		if err != nil {
			log.Println("Upload Cloudinary error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload ảnh thất bại"})
			return
		}
		if imageURL != "" {
			branch.ImageURL = imageURL
		}
	}

	// Cập nhật LastUpdatedAt
	branch.LastUpdatedAt = time.Now()

	// Lưu vào database
	if err := database.DB.Save(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật branch thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch updated successfully", "data": branch})
}
