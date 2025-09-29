package main

import (
	"log"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"time"

	"github.com/robfig/cron/v3"
)

// Task 2: Unlock seat hết hạn
func AutoUnlockSeats() {
	cutoff := time.Now().Add(-3 * time.Minute)
	result := database.DB.
		Model(&models.ShowtimeSeat{}).
		Where("Status = ? AND LockedAt <= ?", 1, cutoff).
		Updates(map[string]interface{}{
			"Status":   0,
			"LockedBy": nil,
		})

	if result.Error != nil {
		log.Printf("[AutoUnlockExpiredSeats] error: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("[AutoUnlockExpiredSeats] auto-unlocked %d seat(s)", result.RowsAffected)
	}
}

// Task 3: Daily update movies
func DailyUpdateMovies() {
	today := time.Now().Truncate(24 * time.Hour)

	result1 := database.DB.Model(&models.Movie{}).
		Where("Status = ? AND DATE(ReleaseDate) = ?", 0, today).
		Update("Status", 1)
	if result1.Error != nil {
		log.Printf("[DailyUpdateMovies] error updating start movies: %v", result1.Error)
	} else if result1.RowsAffected > 0 {
		log.Printf("[DailyUpdateMovies] %d movie(s) set to 'Đang chiếu'", result1.RowsAffected)
	}

	result2 := database.DB.Model(&models.Movie{}).
		Where("Status <> ? AND DATE(LastScreenDate) < ?", 2, today).
		Update("Status", 2)
	if result2.Error != nil {
		log.Printf("[DailyUpdateMovies] error updating end movies: %v", result2.Error)
	} else if result2.RowsAffected > 0 {
		log.Printf("[DailyUpdateMovies] %d movie(s) set to 'Ngưng chiếu'", result2.RowsAffected)
	}
}

// Task 4: Tự động đóng showtime
func AutoCloseShowtimes() {
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
	} else if result.RowsAffected > 0 {
		log.Printf("[AutoCloseShowtimes] %d showtime(s) set to 'Đã chiếu'", result.RowsAffected)
	}
}

// Hàm tạo cron jobs nếu muốn chạy liên tục
func SetupCronJobs() *cron.Cron {
	c := cron.New()

	c.AddFunc("@every 2m", AutoUnlockSeats)
	c.AddFunc("@daily", DailyUpdateMovies)
	c.AddFunc("@every 1m", AutoCloseShowtimes)

	return c
}
