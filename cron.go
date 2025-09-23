package main

import (
	"fmt"
	"log"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"time"

	"github.com/robfig/cron/v3"
)

func SetupCronJobs() *cron.Cron {
	c := cron.New()

	c.AddFunc("@daily", func() {
		fmt.Println("🎬 Cập nhật trạng thái phim...")

		database.DB.Exec(`
        UPDATE movies
        SET Status = CASE
            WHEN Status = 0 AND ReleaseDate <= CURDATE() THEN 1   -- chuyển sang đang chiếu
            WHEN Status != 2 AND LastScreenDate < CURDATE() THEN 2 -- chuyển sang ngừng chiếu
            ELSE Status
        END
    `)
	})

	c.AddFunc("@every 2m", func() {
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
	})

	c.AddFunc("@daily", func() {
		today := time.Now().Truncate(24 * time.Hour) // cắt giờ, chỉ so sánh ngày

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
	})

	c.AddFunc("@every 1m", func() {
		now := time.Now()
		today := now.Format("2006-01-02")  // yyyy-mm-dd
		currentTime := now.Format("15:04") // HH:MM

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
	})

	return c
}
