package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"movie-ticket-booking/config"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"movie-ticket-booking/services"
	"movie-ticket-booking/utils"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(c *gin.Context) {
	var user models.Account

	// Bind JSON input to user
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check for existing email
	var existingUser models.Account
	if err := database.DB.Where("Email = ?", user.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email đã tồn tại"})
		return
	}

	// Set default values
	if user.AccountTypeID == 0 {
		user.AccountTypeID = 1 // Default to 1 if not provided
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Set timestamps
	user.CreatedAt = time.Now()
	user.LastUpdatedAt = time.Now()

	// Save user to database
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng ký thành công",
		"user": gin.H{
			"AccountID":     user.AccountID,
			"AccountTypeID": user.AccountTypeID,
			"BranchID":      user.BranchID,
			"Email":         user.Email,
			"FullName":      user.FullName,
			"PhoneNumber":   user.PhoneNumber,
			"BirthDate":     user.BirthDate,
			"Status":        user.Status,
			"FromFaceBook":  user.FromFacebook,
			"CreatedAt":     user.CreatedAt,
		},
	})
}

type Claims struct {
	AccountID int    `json:"AccountID"`
	Email     string `json:"Email"`
	jwt.RegisteredClaims
}

func LoginUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.Account
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email chưa được đăng ký"})
		return
	}

	if !user.Status {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tài khoản đang bị khóa"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mật khẩu không đúng"})
		return
	}

	// Lấy thông tin chi nhánh
	var branchName string
	if user.BranchID != nil {
		var branch models.Branch
		if err := database.DB.Select("BranchName").Where("BranchID = ?", *user.BranchID).First(&branch).Error; err == nil {
			branchName = branch.BranchName
		}
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		AccountID: user.AccountID,
		Email:     user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.GetJWTKey())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Cấu trúc trả về có thêm BranchName
	type ResponseUser struct {
		AccountID     int       `json:"AccountID"`
		AccountTypeID int       `json:"AccountTypeID"`
		BranchID      *int      `json:"BranchID"`
		BranchName    string    `json:"BranchName,omitempty"`
		Email         string    `json:"Email"`
		PhoneNumber   string    `json:"PhoneNumber"`
		FullName      string    `json:"FullName"`
		BirthDate     string    `json:"BirthDate"`
		Status        bool      `json:"Status"`
		FromFacebook  bool      `json:"FromFacebook"`
		CreatedAt     time.Time `json:"CreatedAt"`
		LastUpdatedAt time.Time `json:"LastUpdatedAt"`
	}

	responseUser := ResponseUser{
		AccountID:     user.AccountID,
		AccountTypeID: user.AccountTypeID,
		BranchID:      user.BranchID,
		BranchName:    branchName,
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		FullName:      user.FullName,
		BirthDate:     user.BirthDate,
		Status:        user.Status,
		FromFacebook:  user.FromFacebook,
		CreatedAt:     user.CreatedAt,
		LastUpdatedAt: user.LastUpdatedAt,
	}

	// Trả về token và thông tin người dùng
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"token":   tokenString,
		"user":    responseUser,
	})
}

func CheckAuth(c *gin.Context) {
	accountID, _ := c.Get("AccountID")
	email, _ := c.Get("Email")

	// Nếu không có user trong context thì unauthorized
	if accountID == nil || email == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Trả về thông tin user
	c.JSON(http.StatusOK, gin.H{
		"AccountID": accountID,
		"Email":     email,
	})
}

func TMidd(c *gin.Context) {
	AccountID, exists := c.Get("AccountID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account ID not found"})
		return
	}

	Email, exists := c.Get("Email")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully accessed protected route",
		"AccountID": AccountID,
		"Email":     Email,
	})
}

func FacebookLogin(c *gin.Context) {
	fbOauthConfig := config.GetFacebookConfig()
	url := fbOauthConfig.AuthCodeURL("random-state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func FacebookCallback(c *gin.Context) {
	fbOauthConfig := config.GetFacebookConfig()
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	token, err := fbOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := fbOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var fbUser struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fbUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	frontendURL := os.Getenv("FE_REACTJS")

	// Kiểm tra tài khoản đã tồn tại chưa
	var user models.Account
	if err := database.DB.Where("email = ?", fbUser.Email).First(&user).Error; err != nil {
		// Nếu chưa, tạo user mới
		user = models.Account{
			Email:         fbUser.Email,
			FullName:      fbUser.Name,
			Password:      "", // không cần mật khẩu cho tài khoản Facebook
			AccountTypeID: 1,
			Status:        true,
			FromFacebook:  true,
			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user from Facebook"})
			return
		}
	} else {
		if !user.Status {
			html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html lang="vi">
			<head>
				<meta charset="UTF-8">
				<title>Tài khoản bị khóa</title>
			</head>
			<body style="text-align:center; margin-top:100px;">
				<h2 style="color:red;">Tài khoản đang bị khóa, liên hệ hotline CSKH để được hỗ trợ: 1099 1889</h2>
				<a href="%s">
					<button style="padding:10px 20px; margin-top:20px;">Về trang chủ</button>
				</a>
			</body>
			</html>
			`, frontendURL)
			c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte(html))
			return
		}
	}

	// Tạo JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		AccountID: user.AccountID,
		Email:     user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.GetJWTKey())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	redirectURL := fmt.Sprintf(
		"%s?token=%s&accountID=%d&accountTypeID=%d&email=%s&fullName=%s&fromFacebook=%t",
		frontendURL,
		url.QueryEscape(tokenString),
		user.AccountID,
		user.AccountTypeID,
		url.QueryEscape(user.Email),
		url.QueryEscape(user.FullName),
		user.FromFacebook,
	)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func GoogleLogin(c *gin.Context) {
	googleOauthConfig := config.GetGoogleConfig()
	url := googleOauthConfig.AuthCodeURL("random-state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {
	googleOauthConfig := config.GetGoogleConfig()
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	// Đổi code lấy access token
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("❌ Google OAuth Exchange error: %v", err) // log ra terminal
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to exchange token",
			"details": err.Error(), // trả chi tiết về cho frontend
		})
		return
	}

	// Gọi Google API để lấy user info
	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?fields=email,name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		Email string `json:"Email"`
		Name  string `json:"Name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	frontendURL := os.Getenv("FE_REACTJS")

	// Kiểm tra trong DB
	var user models.Account
	if err := database.DB.Where("email = ?", googleUser.Email).First(&user).Error; err != nil {
		user = models.Account{
			Email:         googleUser.Email,
			FullName:      googleUser.Name,
			Password:      "", // Google không cần mật khẩu
			AccountTypeID: 1,
			Status:        true,
			FromGoogle:    true,
			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user from Google"})
			return
		}
	} else {
		if !user.Status {
			html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html lang="vi">
			<head>
				<meta charset="UTF-8">
				<title>Tài khoản bị khóa</title>
			</head>
			<body style="text-align:center; margin-top:100px;">
				<h2 style="color:red;">Tài khoản đang bị khóa, liên hệ hotline CSKH để được hỗ trợ: 1099 1889</h2>
				<a href="%s">
					<button style="padding:10px 20px; margin-top:20px;">Về trang chủ</button>
				</a>
			</body>
			</html>
			`, frontendURL)
			c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte(html))
			return
		}
	}

	// Tạo JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		AccountID: user.AccountID,
		Email:     user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.GetJWTKey())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	redirectURL := fmt.Sprintf(
		"%s?token=%s&accountID=%d&accountTypeID=%d&email=%s&fullName=%s&fromGoogle=%t",
		frontendURL,
		url.QueryEscape(tokenString),
		user.AccountID,
		user.AccountTypeID,
		url.QueryEscape(user.Email),
		url.QueryEscape(user.FullName),
		user.FromGoogle,
	)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

var otpStore = map[string]OTPData{}

type OTPData struct {
	Code      string
	ExpiresAt time.Time
}

func RequestOTP(c *gin.Context) {
	email := c.Param("email")
	isValidEmail := func(email string) bool {
		re := regexp.MustCompile(`^[a-zA-Z0-9._%%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		return re.MatchString(email)
	}

	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email không hợp lệ"})
		return
	}

	otp := utils.GenOTP(6)
	otpStore[email] = OTPData{otp, time.Now().Add(5 * time.Minute)}

	if err := services.SendMailReceiveTicket(email, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi OTP"})
}

func VerifyOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	data, ok := otpStore[req.Email]
	if !ok || time.Now().After(data.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP không tồn tại hoặc hết hạn"})
		return
	}
	if req.Code != data.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP không chính xác"})
		return
	}

	delete(otpStore, req.Email)
	c.JSON(http.StatusOK, gin.H{
		"message": "Xác thực thành công",
		"email":   req.Email,
	})
}
