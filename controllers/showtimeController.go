package controllers

import (
	"errors"
	"fmt"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllShowtimesOfDate(c *gin.Context) {
	MovieID := c.Param("MovieID")
	Date := c.Query("showdate")

	type ShowtimeRow struct {
		BranchName string
		ShowtimeID int
		StartTime  string
	}

	var rows []ShowtimeRow

	err := database.DB.Raw(`
        SELECT 
            b.BranchName,
            s.ShowtimeID,
            s.StartTime
        FROM showtimes s
        JOIN theaters t ON s.TheaterID = t.TheaterID
        JOIN branches b ON t.BranchID = b.BranchID
        WHERE s.MovieID = ? 
          AND DATE(s.ShowDate) = ? 
          AND s.IsOpenOrder = 1
		  AND s.Status = 1
        ORDER BY b.BranchName, s.StartTime
    `, MovieID, Date).Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve showtimes"})
		return
	}

	if len(rows) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No showtimes found for this movie on selected date"})
		return
	}

	type Showtime struct {
		ShowtimeID int    `json:"ShowtimeID"`
		StartTime  string `json:"StartTime"`
	}
	type BranchGroup struct {
		BranchName string     `json:"BranchName"`
		Showtimes  []Showtime `json:"Showtimes"`
	}

	var result []BranchGroup

	for _, r := range rows {
		// tìm branch đã có trong result
		found := false
		for i := range result {
			if result[i].BranchName == r.BranchName {
				result[i].Showtimes = append(result[i].Showtimes, Showtime{
					ShowtimeID: r.ShowtimeID,
					StartTime:  r.StartTime,
				})
				found = true
				break
			}
		}
		// nếu chưa có thì thêm mới
		if !found {
			result = append(result, BranchGroup{
				BranchName: r.BranchName,
				Showtimes: []Showtime{{
					ShowtimeID: r.ShowtimeID,
					StartTime:  r.StartTime,
				}},
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func GetShowtimeInfo(c *gin.Context) {
	// Lấy movie ID từ tham số URL
	showtimeid := c.Param("ShowtimeID")

	// Khai báo struct để nhận dữ liệu từ truy vấn
	type ShowtimeResponse struct {
		MovieID     int
		ShowtimeID  int
		MovieName   string
		Poster      string
		Duration    int
		AgeTag      string
		BranchID    int
		BranchName  string
		TheaterName string
		StartTime   string
		ShowDate    string
	}

	var showtimes []ShowtimeResponse

	// Thực hiện truy vấn
	err := database.DB.Raw(`
		SELECT 
			m.MovieID,
    		s.ShowtimeID,
    		m.MovieName,
    		m.Poster,
    		m.Duration,
			m.AgeTag,
			b.BranchID,
    		b.BranchName,
    		t.TheaterName,
    		s.StartTime,
    		s.ShowDate
		FROM showtimes s
		JOIN theaters t ON s.TheaterID = t.TheaterID
		JOIN branches b ON t.BranchID = b.BranchID
		JOIN movies m ON s.MovieID = m.MovieID 
		WHERE s.ShowtimeID = ?`, showtimeid).Scan(&showtimes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve showtimes"})
		return
	}

	// Kiểm tra nếu không có dữ liệu
	if len(showtimes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No showtimes found for this movie"})
		return
	}

	// Trả về danh sách showtime
	c.JSON(http.StatusOK, gin.H{"data": showtimes})
}

func GetAllShowtimesOfBranch(c *gin.Context) {
	// Lấy BranchID từ tham số URL
	branchID := c.Param("BranchID")
	showDate := c.Query("ShowDate") // Lấy giá trị ShowDate từ query parameter

	// Khai báo struct để nhận dữ liệu từ truy vấn
	type ShowtimeResponse struct {
		ShowtimeID   int
		TheaterID    int
		Poster       string
		TheaterName  string
		StartTime    string
		EndTime      string
		ShowDate     string
		Status       int
		IsOpenOrder  bool
		CancelReason string
	}

	var showtimes []ShowtimeResponse

	// Xây dựng câu truy vấn SQL
	query := `
		SELECT 
			s.ShowtimeID,
			t.TheaterID,
			m.Poster,
			t.TheaterName,
			s.StartTime,
			s.EndTime,
			s.ShowDate,
			s.Status,
			s.IsOpenOrder,
			s.CancelReason
		FROM showtimes s
		JOIN theaters t ON s.TheaterID = t.TheaterID
		JOIN branches b ON t.BranchID = b.BranchID
		JOIN movies m ON s.MovieID = m.MovieID
		WHERE b.BranchID = ?`

	// Nếu có ShowDate, thêm điều kiện lọc theo ngày
	var params []interface{}
	params = append(params, branchID)
	if showDate != "" {
		query += " AND s.ShowDate = ?"
		params = append(params, showDate)
	}

	// Thực hiện truy vấn
	err := database.DB.Raw(query, params...).Scan(&showtimes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve showtimes"})
		return
	}

	// Kiểm tra nếu không có dữ liệu
	// if len(showtimes) == 0 {
	// 	c.JSON(http.StatusNotFound, gin.H{"message": "No showtimes found for this branch and date"})
	// 	return
	// }

	// Trả về danh sách showtimes
	c.JSON(http.StatusOK, gin.H{"data": showtimes})
}

func AddShowtime(c *gin.Context) {
	// Khai báo struct request
	type CreateShowtimeRequest struct {
		TheaterID int    `json:"TheaterID"`
		MovieID   int    `json:"MovieID"`
		ShowDate  string `json:"ShowDate"` // format YYYY-MM-DD
		StartTime string `json:"StartTime"`
		EndTime   string `json:"EndTime"`
		Status    int    `json:"Status"`
		CreatedBy string `json:"CreatedBy"`
	}

	var request CreateShowtimeRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ✅ Parse ngày suất chiếu
	dateLayout := "2006-01-02"
	showDate, err := time.Parse(dateLayout, request.ShowDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ShowDate format, must be YYYY-MM-DD"})
		return
	}

	// ✅ Lấy ngày hôm nay (0h00)
	today := time.Now().Truncate(24 * time.Hour)

	// Nếu ngày suất chiếu <= hôm nay => lỗi
	if !showDate.After(today) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Đã chốt suất chiếu trong ngày, chỉ có thể thêm suất chiếu vào những ngày chưa đến"})
		return
	}

	// ✅ Lấy thông tin phim
	var movie models.Movie
	if err := database.DB.First(&movie, request.MovieID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy phim"})
		return
	}

	// Parse ReleaseDate và LastScreenDate
	releaseDate, err := time.Parse(dateLayout, movie.ReleaseDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi parse ReleaseDate của phim"})
		return
	}
	lastScreenDate, err := time.Parse(dateLayout, movie.LastScreenDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi parse LastScreenDate của phim"})
		return
	}

	// Check ngày suất chiếu so với phim
	if showDate.Before(releaseDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chưa đến ngày khởi chiếu"})
		return
	}
	if showDate.After(lastScreenDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phim này đã kết thúc chiếu rạp"})
		return
	}

	// ✅ Parse giờ chiếu
	layout := "15:04"
	newStart, err := time.Parse(layout, request.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid StartTime format, must be HH:mm"})
		return
	}
	newEnd, err := time.Parse(layout, request.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid EndTime format, must be HH:mm"})
		return
	}
	if !newEnd.After(newStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "EndTime phải lớn hơn StartTime"})
		return
	}

	// ✅ Lấy tất cả suất chiếu cùng ngày & cùng rạp
	var existingShowtimes []models.Showtime
	if err := database.DB.Where("ShowDate = ? AND TheaterID = ?", request.ShowDate, request.TheaterID).
		Find(&existingShowtimes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// ✅ Kiểm tra chồng chéo & khoảng cách 10 phút trong cùng rạp
	for _, s := range existingShowtimes {
		exStart, _ := time.Parse(layout, s.StartTime)
		exEnd, _ := time.Parse(layout, s.EndTime)

		// Nếu khoảng [newStart, newEnd] giao với [exStart, exEnd]
		if newStart.Before(exEnd) && newEnd.After(exStart) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Suất chiếu trùng giờ với suất %s - %s", s.StartTime, s.EndTime),
			})
			return
		}

		// Nếu suất mới bắt đầu trong vòng 10 phút sau khi suất cũ kết thúc
		if newStart.After(exEnd) && newStart.Before(exEnd.Add(10*time.Minute)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Suất chiếu mới phải cách ít nhất 10 phút sau suất %s - %s", s.StartTime, s.EndTime),
			})
			return
		}

		// Nếu suất mới kết thúc trong vòng 10 phút trước khi suất cũ bắt đầu
		if newEnd.Before(exStart) && newEnd.After(exStart.Add(-10*time.Minute)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Suất chiếu mới phải kết thúc ít nhất 10 phút trước suất %s - %s", s.StartTime, s.EndTime),
			})
			return
		}
	}

	// ✅ Lấy branchId từ Theater
	var theater models.Theater
	if err := database.DB.First(&theater, request.TheaterID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy rạp"})
		return
	}
	branchID := theater.BranchID

	// ✅ Lấy tất cả suất chiếu cùng ngày & cùng branch & cùng phim
	var sameMovieShowtimes []struct {
		models.Showtime
		TheaterName string
	}
	if err := database.DB.
		Model(&models.Showtime{}).
		Select("showtimes.*, t.TheaterName").
		Joins("JOIN theaters t ON t.TheaterID = showtimes.TheaterID").
		Where("showtimes.ShowDate = ? AND t.BranchID = ? AND showtimes.MovieID = ?", request.ShowDate, branchID, request.MovieID).
		Scan(&sameMovieShowtimes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// ✅ Check StartTime cách nhau >= 30 phút giữa các rạp trong branch
	for _, s := range sameMovieShowtimes {
		exStart, _ := time.Parse(layout, s.StartTime)
		diff := newStart.Sub(exStart)
		if diff < 0 {
			diff = -diff
		}

		if diff < 30*time.Minute {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"Suất chiếu của phim này phải cách ít nhất 30 phút so với suất %s ở %s",
					s.StartTime, s.TheaterName,
				),
			})
			return
		}
	}

	// ✅ Nếu hợp lệ -> thêm mới
	showtime := models.Showtime{
		TheaterID: request.TheaterID,
		MovieID:   request.MovieID,
		ShowDate:  request.ShowDate,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
		Status:    request.Status,
		CreatedBy: request.CreatedBy,
	}

	if err := database.DB.Create(&showtime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create showtime"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": showtime})
}

func GetDetailsShowtime(c *gin.Context) {
	var showtime models.Showtime
	showtimeID := c.Param("ShowtimeID") // Lấy ID từ URL

	// Tìm suất chiếu kèm thông tin phim và rạp chiếu
	if err := database.DB.Preload("Movie").Preload("Theater").
		Where("ShowtimeID = ?", showtimeID).
		First(&showtime).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
		return
	}

	// Trả về dữ liệu suất chiếu
	c.JSON(http.StatusOK, gin.H{"data": showtime})
}

func OpenOrderShowtime(c *gin.Context) {
	id := c.Param("ShowtimeID")

	// Tìm suất chiếu
	var showtime models.Showtime
	if err := database.DB.First(&showtime, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find showtime"})
		return
	}

	// Parse ShowDate + StartTime => time.Time
	layout := "2006-01-02 15:04"
	showtimeStr := fmt.Sprintf("%s %s", showtime.ShowDate, showtime.StartTime)
	showtimeStart, err := time.ParseInLocation(layout, showtimeStr, time.Local)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid showtime format"})
		return
	}

	// Deadline = StartTime - 2h
	deadline := showtimeStart.Add(-2 * time.Hour)
	now := time.Now()

	if now.After(deadline) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"Hiện tại đã là %s, không thể mở đặt vé đối với suất chiếu trước %s trong hôm nay",
				now.Format("15:04"),
				now.Add(2*time.Hour).Format("15:04"),
			),
		})
		return
	}

	// Nếu đã mở rồi thì không mở lại
	if showtime.IsOpenOrder {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Showtime already opened for order"})
		return
	}

	// ✅ Lấy email từ context
	email := c.GetString("Email")
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - missing email"})
		return
	}

	// ✅ Cập nhật IsOpenOrder + LastUpdatedBy
	if err := database.DB.Model(&showtime).Updates(map[string]interface{}{
		"IsOpenOrder":   true,
		"LastUpdatedBy": email,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open order for showtime"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mở đặt vé thành công"})
}

func CancelShowtime(c *gin.Context) {
	// Lấy ShowtimeID từ URL
	id := c.Param("ShowtimeID")

	// Bind dữ liệu từ body
	var body struct {
		CancelReason string `json:"CancelReason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Tìm suất chiếu theo ID
	var showtime models.Showtime
	if err := database.DB.First(&showtime, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find showtime"})
		return
	}

	// Lấy Email từ context (được gán trong middleware)
	emailRaw, exists := c.Get("Email")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found in context"})
		return
	}

	email, ok := emailRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email is not a string"})
		return
	}

	// Cập nhật Status = 2 (hủy), CancelReason, LastUpdatedAt, LastUpdatedBy
	updateData := map[string]interface{}{
		"Status":        2,
		"CancelReason":  body.CancelReason,
		"LastUpdatedAt": time.Now(),
		"LastUpdatedBy": email,
	}

	if err := database.DB.Model(&showtime).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel showtime"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Hủy suất chiếu thành công",
		"cancel_reason":   body.CancelReason,
		"last_updated_by": email,
	})
}

func DeleteShowtime(c *gin.Context) {
	// Get showtime ID from the URL parameter
	id := c.Param("ShowtimeID")

	// Find the showtime by ID
	var showtime models.Showtime
	if err := database.DB.First(&showtime, id).Error; err != nil {
		if gorm.ErrRecordNotFound == err {
			c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find showtime"})
		return
	}

	// Delete the showtime
	if err := database.DB.Delete(&showtime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete showtime"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Showtime deleted successfully"})
}
