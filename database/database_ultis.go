package database

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DisableForeignKeyChecks tắt kiểm tra khóa ngoại trong database
func DisableForeignKeyChecks(db *gorm.DB, c *gin.Context) error {
	if err := db.Exec("SET foreign_key_checks = 0;").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable foreign key checks"})
		return err
	}
	return nil
}
