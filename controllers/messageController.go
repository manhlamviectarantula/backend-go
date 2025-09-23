package controllers

import (
	"fmt"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SendMessage(c *gin.Context) {
	senderIDRaw, exists := c.Get("AccountID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account ID not found"})
		return
	}

	senderID, ok := senderIDRaw.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account ID is not an integer"})
		return
	}

	receiverIDStr := c.Param("userToChatID")
	text := c.PostForm("Text")

	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ReceiverID không hợp lệ"})
		return
	}

	// Tạo struct Message
	message := models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Text:       text,
	}

	// Lưu vào database
	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể gửi tin nhắn", "details": err.Error()})
		return
	}

	// Phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Gửi tin nhắn thành công",
		"data":    message,
	})
}

func GetMessages(c *gin.Context) {
	myID, exists := c.Get("AccountID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account ID not found"})
		return
	}

	myIDInt, ok := myID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account ID is not an integer"})
		return
	}

	userToChatID := c.Param("userToChatID")

	chatIDInt, err := strconv.Atoi(userToChatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người cần chat không hợp lệ"})
		return
	}

	// Truy vấn tin nhắn
	var messages []models.Message
	if err := database.DB.Where(
		"(SenderID = ? AND ReceiverID = ?) OR (SenderID = ? AND ReceiverID = ?)",
		myIDInt, chatIDInt, chatIDInt, myIDInt,
	).Order("MessageID asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy tin nhắn", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func GetUsersSidebarAdmin(c *gin.Context) {
	// Lấy adminID từ context (middleware đã gán)
	adminID, exists := c.Get("AccountID")
	fmt.Println("AccountID in context:", adminID) // ✅ Dùng lại biến đã có
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin ID not found"})
		return
	}

	adminIDInt, ok := adminID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin ID is not an integer"})
		return
	}

	// Subquery: lấy tất cả SenderID trong bảng messages mà ReceiverID là admin
	subQuery := database.DB.Model(&models.Message{}).
		Select("SenderID").
		Where("ReceiverID = ?", adminIDInt).
		Group("SenderID")

	// Truy vấn bảng accounts để lấy thông tin người dùng
	var users []models.Account
	if err := database.DB.Where("AccountID IN (?)", subQuery).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách người dùng", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
