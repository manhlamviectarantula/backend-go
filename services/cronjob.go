package services

import (
	"log"
	"net/http"
	"time"

	"movie-ticket-booking/database"
	"movie-ticket-booking/models"

	"github.com/gin-gonic/gin"
)

// -------------------- Task 1: Daily update movies --------------------
func DailyUpdateMoviesHandler(c *gin.Context) {
	today := time.Now().Truncate(24 * time.Hour)

	// Update start movies
	result1 := database.DB.Model(&models.Movie{}).
		Where("Status = ? AND DATE(ReleaseDate) = ?", 0, today).
		Update("Status", 1)
	if result1.Error != nil {
		log.Printf("[DailyUpdateMovies] error updating start movies: %v", result1.Error)
	}

	// Update end movies
	result2 := database.DB.Model(&models.Movie{}).
		Where("Status <> ? AND DATE(LastScreenDate) < ?", 2, today).
		Update("Status", 2)
	if result2.Error != nil {
		log.Printf("[DailyUpdateMovies] error updating end movies: %v", result2.Error)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "DailyUpdateMovies executed",
		"start_movies": result1.RowsAffected,
		"end_movies":   result2.RowsAffected,
	})
}

// -------------------- Task 2: Unlock seat hết hạn --------------------
func AutoUnlockSeatsHandler(c *gin.Context) {
	cutoff := time.Now().Add(-3 * time.Minute)

	result := database.DB.Model(&models.ShowtimeSeat{}).
		Where("Status = ? AND LockedAt <= ?", 1, cutoff).
		Updates(map[string]interface{}{
			"Status":   0,
			"LockedBy": nil,
		})

	if result.Error != nil {
		log.Printf("[AutoUnlockSeats] error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "AutoUnlockSeats executed",
		"rows_affected": result.RowsAffected,
	})
}

// -------------------- Task 3: Tự động đóng showtime --------------------
func AutoCloseShowtimesHandler(c *gin.Context) {
	now := time.Now()
	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")

	result := database.DB.Model(&models.Showtime{}).
		Where("Status = 1").
		Where("IsOpenOrder = ?", true).
		Where("(ShowDate < ?) OR (ShowDate = ? AND StartTime < ?)", today, today, currentTime).
		Update("Status", 0)

	if result.Error != nil {
		log.Printf("[AutoCloseShowtimes] error updating showtimes: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "AutoCloseShowtimes executed",
		"rows_affected": result.RowsAffected,
	})
}
