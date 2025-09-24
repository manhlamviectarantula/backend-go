package controllers

import (
	"fmt"
	"html"
	"log"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"movie-ticket-booking/services"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UpdateFood(c *gin.Context) {
	foodID, err := strconv.Atoi(c.Param("FoodID"))
	if err != nil {
		log.Println("Invalid FoodID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FoodID"})
		return
	}

	var food models.Food
	if err := database.DB.First(&food, foodID).Error; err != nil {
		log.Println("Food not found:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Món ăn không tồn tại"})
		return
	}

	// Bắt buộc phải có LastUpdatedBy
	lastUpdatedBy := strings.TrimSpace(c.PostForm("LastUpdatedBy"))
	if lastUpdatedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LastUpdatedBy là bắt buộc"})
		return
	}
	food.LastUpdatedBy = html.EscapeString(lastUpdatedBy)

	// Mở transaction
	tx := database.DB.Begin()

	// Xử lý upload file ảnh (nếu có)
	imageFile, err := c.FormFile("Image")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(imageFile.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Chỉ chấp nhận file ảnh .jpg, .jpeg, .png"})
			return
		}

		newFileName := fmt.Sprintf("upload/%s%s", uuid.New().String(), ext)
		if err := c.SaveUploadedFile(imageFile, newFileName); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu file ảnh thất bại"})
			return
		}

		// Xóa ảnh cũ nếu có
		if food.Image != "" {
			if _, err := os.Stat(food.Image); err == nil {
				os.Remove(food.Image)
			}
		}
		food.Image = newFileName
	}

	// Cập nhật các thông tin khác (nếu có)
	if foodName := strings.TrimSpace(c.PostForm("FoodName")); foodName != "" {
		food.FoodName = html.EscapeString(foodName)
	}
	if description := strings.TrimSpace(c.PostForm("Description")); description != "" {
		food.Description = html.EscapeString(description)
	}
	if priceStr := c.PostForm("Price"); priceStr != "" {
		if price, err := strconv.Atoi(priceStr); err == nil {
			food.Price = price
		} else {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Price phải là số nguyên hợp lệ"})
			return
		}
	}

	// GORM sẽ tự động cập nhật LastUpdatedAt
	if err := tx.Save(&food).Error; err != nil {
		log.Println("Database error:", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật món ăn"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật món ăn thành công", "data": food})
}

func AddFoodOfBranch(c *gin.Context) {
	// Chuyển BranchID từ string sang int
	branchID, err := strconv.Atoi(c.Param("BranchID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "BranchID không hợp lệ"})
		return
	}

	// Lấy file ảnh
	imageFile, err := c.FormFile("Image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ảnh là bắt buộc"})
		return
	}

	// Upload lên Cloudinary
	imageURL, err := services.UploadToCloudinary(imageFile, "foods")
	if err != nil {
		log.Println("Upload Cloudinary error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload ảnh thất bại"})
		return
	}

	if imageURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload ảnh thất bại, URL rỗng"})
		return
	}

	// Chuyển Price sang int
	priceStr := c.PostForm("Price")
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price phải là số nguyên"})
		return
	}

	food := models.Food{
		BranchID:      branchID,
		FoodName:      c.PostForm("FoodName"),
		Image:         imageURL, // dùng URL từ Cloudinary
		Description:   c.PostForm("Description"),
		Price:         price,
		CreatedBy:     c.PostForm("CreatedBy"),
		LastUpdatedBy: c.PostForm("LastUpdatedBy"),
	}

	// Thêm món ăn vào database
	if err := database.DB.Create(&food).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tạo food thất bại"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Food added successfully", "data": food})
}

func GetFoodsOfBranch(c *gin.Context) {
	var foods []models.Food
	branchID := c.Param("BranchID") // Lấy BranchID từ URL

	// Truy vấn cơ sở dữ liệu lấy danh sách món ăn của chi nhánh
	result := database.DB.Where("BranchID = ?", branchID).Find(&foods)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	// Trả về danh sách món ăn
	c.JSON(200, foods)
}

func DeleteFood(c *gin.Context) {
	foodID, err := strconv.Atoi(c.Param("FoodID"))
	if err != nil {
		log.Println("Invalid FoodID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FoodID"})
		return
	}

	var food models.Food
	if err := database.DB.First(&food, foodID).Error; err != nil {
		log.Println("Food not found:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Món ăn không tồn tại"})
		return
	}

	// Mở transaction
	tx := database.DB.Begin()

	// Xóa ảnh nếu có
	if food.Image != "" {
		if _, err := os.Stat(food.Image); err == nil {
			if err := os.Remove(food.Image); err != nil {
				log.Println("Lỗi khi xóa ảnh:", err)
				// Không rollback vì lỗi này không ảnh hưởng đến DB
			}
		}
	}

	// Xóa món ăn khỏi DB
	if err := tx.Delete(&food).Error; err != nil {
		log.Println("Lỗi khi xóa món ăn:", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa món ăn"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Xóa món ăn thành công"})
}
