package controllers

import (
	"log"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"movie-ticket-booking/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllMovies(c *gin.Context) {
	var movies []models.Movie

	// Query all movies from the database
	if err := database.DB.Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	// Return the list of movies
	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

func GetMoviesInAddShowtime(c *gin.Context) {
	var movies []models.Movie

	// Lấy phim có Status = 0 hoặc 1
	if err := database.DB.
		Where("Status IN ?", []int{0, 1}).
		Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

func AddMovie(c *gin.Context) {
	posterFile, err := c.FormFile("Poster")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Poster là bắt buộc"})
		return
	}

	posterURL, err := services.UploadToCloudinary(posterFile, "movies")
	if err != nil {
		log.Println("Upload Cloudinary error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload poster thất bại"})
		return
	}

	if posterURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload poster thất bại, URL rỗng"})
		return
	}

	durationStr := c.PostForm("Duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Duration phải là số nguyên"})
		return
	}

	ratingStr := c.PostForm("Rating")
	rating, err := strconv.ParseFloat(ratingStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating phải là số thập phân"})
		return
	}

	movie := models.Movie{
		MovieName:      c.PostForm("MovieName"),
		Slug:           c.PostForm("Slug"),
		AgeTag:         c.PostForm("AgeTag"),
		Duration:       duration,
		ReleaseDate:    c.PostForm("ReleaseDate"),
		LastScreenDate: c.PostForm("LastScreenDate"),
		Poster:         posterURL,
		Trailer:        c.PostForm("Trailer"),
		Rating:         rating,
		Description:    c.PostForm("Description"),
		CreatedBy:      c.PostForm("CreatedBy"),
		LastUpdatedBy:  c.PostForm("LastUpdatedBy"),
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tạo movie thất bại"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": movie})
}

func GetShowingMovie(c *gin.Context) {
	var movies []models.Movie

	// Truy vấn tất cả phim có status = 1
	if err := database.DB.Where("Status = ?", 1).Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách phim"})
		return
	}

	// Trả về danh sách phim
	c.JSON(http.StatusOK, gin.H{"data": movies})
}

func GetUpcomingMovie(c *gin.Context) {
	var movies []models.Movie

	// Truy vấn tất cả phim có status = 0
	if err := database.DB.Where("Status = ?", 0).Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách phim"})
		return
	}

	// Trả về danh sách phim
	c.JSON(http.StatusOK, gin.H{"data": movies})
}

func UpdateMovie(c *gin.Context) {
	// Lấy movieID từ URL
	movieID, err := strconv.Atoi(c.Param("MovieID"))
	if err != nil {
		log.Println("Invalid MovieID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MovieID"})
		return
	}

	var movie models.Movie
	// Tìm bộ phim trong cơ sở dữ liệu
	if err := database.DB.First(&movie, movieID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Kiểm tra xem có file poster mới được upload không
	posterFile, err := c.FormFile("Poster")
	if err == nil { // Nếu có file mới
		filePath := "upload/" + posterFile.Filename
		if err := c.SaveUploadedFile(posterFile, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu file poster thất bại"})
			return
		}
		movie.Poster = filePath
	}

	// Chuyển đổi Duration thành int
	if durationStr := c.Request.FormValue("Duration"); durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duration phải là một số nguyên"})
			return
		}
		movie.Duration = duration
	}

	// Chuyển đổi Rating thành float
	if ratingStr := c.Request.FormValue("Rating"); ratingStr != "" {
		rating, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rating phải là một số thập phân"})
			return
		}
		movie.Rating = rating
	}

	// Chuyển đổi Status thành int
	if statusStr := c.Request.FormValue("Status"); statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status phải là một số nguyên"})
			return
		}
		movie.Status = status
	}

	// Cập nhật các trường dữ liệu nếu có giá trị mới
	if movieName := c.Request.FormValue("MovieName"); movieName != "" {
		movie.MovieName = movieName
	}
	if slug := c.Request.FormValue("Slug"); slug != "" {
		movie.Slug = slug
	}
	if ageTag := c.Request.FormValue("AgeTag"); ageTag != "" {
		movie.AgeTag = ageTag
	}
	if releaseDate := c.Request.FormValue("ReleaseDate"); releaseDate != "" {
		movie.ReleaseDate = releaseDate
	}
	if lastScreenDate := c.Request.FormValue("LastScreenDate"); lastScreenDate != "" {
		movie.LastScreenDate = lastScreenDate
	}
	if trailer := c.Request.FormValue("Trailer"); trailer != "" {
		movie.Trailer = trailer
	}
	if description := c.Request.FormValue("Description"); description != "" {
		movie.Description = description
	}
	if lastUpdatedBy := c.Request.FormValue("LastUpdatedBy"); lastUpdatedBy != "" {
		movie.LastUpdatedBy = lastUpdatedBy
	}

	// Cập nhật bản ghi trong database
	if err := database.DB.Save(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật movie thất bại"})
		return
	}

	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"data": movie})
}

func GetDetailsMovie(c *gin.Context) {
	var movie models.Movie
	id := c.Param("id")

	// Tìm bộ phim theo ID
	if err := database.DB.First(&movie, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Trả về thông tin bộ phim
	c.JSON(http.StatusOK, gin.H{"data": movie})
}
