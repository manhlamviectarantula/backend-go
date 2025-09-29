package controllers

import (
	"fmt"
	"math"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"movie-ticket-booking/services"
	"movie-ticket-booking/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func FindAccountByID(accountID int) (map[string]interface{}, error) {
	var account struct {
		AccountID       int
		Email           string
		PhoneNumber     string
		FullName        string
		BirthDate       string
		Status          int
		CreatedAt       time.Time
		LastUpdatedAt   time.Time
		AccountTypeID   int
		AccountTypeName string
	}

	query := `
	SELECT 
		a.AccountID, 
		a.Email, 
		a.PhoneNumber, 
		a.FullName, 
		a.BirthDate, 
		a.Status, 
		a.CreatedAt, 
		a.LastUpdatedAt, 
		a.AccountTypeID,
		at.AccountTypeName
	FROM accounts a
	JOIN account_types at ON a.AccountTypeID = at.AccountTypeID
	WHERE a.AccountID = ?
	`

	if err := database.DB.Raw(query, accountID).Scan(&account).Error; err != nil {
		return nil, err
	}

	if account.AccountID == 0 {
		return nil, fmt.Errorf("account not found")
	}

	// Trả về dưới dạng map để tiện dùng ở chatbot
	return map[string]interface{}{
		"AccountID":       account.AccountID,
		"Email":           account.Email,
		"PhoneNumber":     account.PhoneNumber,
		"FullName":        account.FullName,
		"BirthDate":       account.BirthDate,
		"Status":          account.Status,
		"CreatedAt":       account.CreatedAt,
		"LastUpdatedAt":   account.LastUpdatedAt,
		"AccountTypeID":   account.AccountTypeID,
		"AccountTypeName": account.AccountTypeName,
	}, nil
}

type AccountResponse struct {
	AccountID       int    `json:"AccountID"`
	Email           string `json:"Email"`
	PhoneNumber     string `json:"PhoneNumber"`
	FullName        string `json:"FullName"`
	BirthDate       string `json:"BirthDate"`
	Status          bool   `json:"Status"`
	CreatedAt       string `json:"CreatedAt"`
	LastUpdatedAt   string `json:"LastUpdatedAt"`
	AccountTypeID   int    `json:"AccountTypeID"`
	AccountTypeName string `json:"AccountTypeName"`
	BranchName      string `json:"BranchName"`
}

func GetAllAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "7"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := database.DB.Model(&models.Account{}).Where("AccountTypeID != ?", 3).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count accounts"})
		return
	}

	// Lấy data với join trực tiếp
	var accounts []AccountResponse
	if err := database.DB.Table("accounts a").
		Select(`
			a.AccountID, 
			a.Email, 
			a.PhoneNumber, 
			a.FullName, 
			a.BirthDate, 
			a.Status, 
			a.CreatedAt, 
			a.LastUpdatedAt, 
			a.AccountTypeID,
			at.AccountTypeName,
			COALESCE(b.BranchName, '') AS BranchName
		`).
		Joins("JOIN account_types at ON a.AccountTypeID = at.AccountTypeID").
		Joins("LEFT JOIN branches b ON a.BranchID = b.BranchID").
		Where("a.AccountTypeID != ?", 3).
		Order("a.CreatedAt DESC").
		Limit(limit).Offset(offset).
		Scan(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       accounts,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": int(math.Ceil(float64(total) / float64(limit))),
	})
}

func BlockAccount(c *gin.Context) {
	var account models.Account

	// Lấy account ID từ URL và kiểm tra hợp lệ
	accountID := c.Param("AccountID")
	id, err := strconv.Atoi(accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	// Tìm tài khoản theo ID
	if err := database.DB.Where("AccountID = ?", id).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Đảo trạng thái tài khoản
	account.Status = !account.Status
	account.LastUpdatedAt = time.Now()

	// Lưu thay đổi vào database
	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account status"})
		return
	}

	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Account status updated successfully",
		"status":  account.Status,
	})
}

func UpdateAccount(c *gin.Context) {
	var account models.Account

	// Get the account ID from the URL
	accountID := c.Param("AccountID")

	// Find the account by ID
	if err := database.DB.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Lưu lại email & phone cũ để so sánh
	oldEmail := account.Email
	oldPhone := account.PhoneNumber

	// Bind incoming JSON to the Account struct
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Kiểm tra nếu email bị thay đổi và đã tồn tại ở tài khoản khác
	if account.Email != oldEmail {
		var existingAccount models.Account
		if err := database.DB.Where("Email = ? AND AccountID != ?", account.Email, account.AccountID).First(&existingAccount).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email đã tồn tại"})
			return
		}
	}

	// Kiểm tra nếu số điện thoại bị thay đổi và đã tồn tại ở tài khoản khác
	if account.PhoneNumber != oldPhone {
		var existingAccount models.Account
		if err := database.DB.Where("PhoneNumber = ? AND AccountID != ?", account.PhoneNumber, account.AccountID).First(&existingAccount).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại đã tồn tại"})
			return
		}
	}

	// Set lại LastUpdatedAt
	account.LastUpdatedAt = time.Now()

	// Cập nhật tài khoản
	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"data": account})
}

func UpdatePassword(c *gin.Context) {
	var input struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	// Get the account ID from the URL
	accountID := c.Param("AccountID")

	// Find the account by ID
	var account models.Account
	if err := database.DB.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Bind incoming JSON to the input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the current password
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai mật khẩu hiện tại"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update the password
	account.Password = string(hashedPassword)
	account.LastUpdatedAt = time.Now()

	// Save the updated account to the database
	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Sửa mật khẩu thành công"})
}

func ForgetPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email không hợp lệ"})
		return
	}

	// Tìm account theo email
	var account models.Account
	if err := database.DB.Where("email = ?", input.Email).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tài khoản với email này"})
		return
	}

	// Sinh mật khẩu ngẫu nhiên
	newPassword := utils.GenerateRandomPassword(10)

	// Hash mật khẩu mới
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không hash được mật khẩu mới"})
		return
	}

	// Cập nhật mật khẩu trong DB
	account.Password = string(hashedPassword)
	account.LastUpdatedAt = time.Now()
	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật mật khẩu thất bại"})
		return
	}

	// Gửi email mật khẩu mới
	if err := services.SendNewPasswordEmail(account.Email, newPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không gửi được email, vui lòng thử lại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mật khẩu mới đã được gửi đến email của bạn"})
}

type UpgradeAccountRequest struct {
	AccountID     int `json:"AccountID" binding:"required"`
	BranchID      int `json:"BranchID" binding:"required"`
	AccountTypeID int `json:"AccountTypeID" binding:"required"`
}

func UpgradeAccount(c *gin.Context) {
	var req UpgradeAccountRequest

	// Bind JSON từ body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Tìm account theo ID
	var account models.Account
	if err := database.DB.First(&account, req.AccountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Cập nhật BranchID và AccountTypeID
	account.BranchID = &req.BranchID
	account.AccountTypeID = req.AccountTypeID

	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account upgraded successfully",
		"account": account,
	})
}

type DowngradeAccountRequest struct {
	AccountID     int `json:"AccountID" binding:"required"`
	AccountTypeID int `json:"AccountTypeID" binding:"required"`
}

func DowngradeAccount(c *gin.Context) {
	var req DowngradeAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Tìm account
	var account models.Account
	if err := database.DB.First(&account, req.AccountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Chỉ downgrade nếu là quản lý chi nhánh
	if account.AccountTypeID != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only branch managers can be downgraded"})
		return
	}

	// Reset BranchID và đổi loại tài khoản về user thường (1)
	account.BranchID = nil
	account.AccountTypeID = req.AccountTypeID

	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to downgrade account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account downgraded successfully",
		"account": account,
	})
}
