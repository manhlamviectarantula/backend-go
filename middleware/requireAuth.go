package middleware

import (
	"movie-ticket-booking/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	AccountID int    `json:"AccountID"`
	Email     string `json:"Email"`
	jwt.RegisteredClaims
}

func RequireLogin(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		c.Abort()
		return
	}

	// Tách Bearer từ "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Malformed token"})
		c.Abort()
		return
	}

	// Parse token và kiểm tra tính hợp lệ
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra phương thức ký của token (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrNoLocation
		}
		return config.GetJWTKey(), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	// Lưu thông tin người dùng vào context nếu token hợp lệ
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		c.Set("AccountID", claims.AccountID)
		c.Set("Email", claims.Email)
	}

	// Tiếp tục với request
	c.Next()
}
