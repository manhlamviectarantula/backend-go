package controllers

import (
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllTheaterOfBranch(c *gin.Context) {
	// Lấy BranchID từ URL
	branchID := c.Param("BranchID")

	var theaters []models.Theater

	// Truy vấn danh sách rạp phim theo BranchID
	if err := database.DB.Where("BranchID = ?", branchID).Find(&theaters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn cơ sở dữ liệu"})
		return
	}

	// Trả về danh sách rạp phim
	c.JSON(http.StatusOK, gin.H{"data": theaters})
}

func GetDetailsTheater(c *gin.Context) {
	// Lấy TheaterID từ URL
	theaterID := c.Param("TheaterID")

	var theater models.Theater

	// Tìm rạp phim theo ID
	if err := database.DB.First(&theater, theaterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy rạp phim"})
		return
	}

	// Trả về thông tin chi tiết của rạp phim
	c.JSON(http.StatusOK, gin.H{"data": theater})
}

func GetSeatsOfTheater(c *gin.Context) {
	theaterID := c.Param("TheaterID") // Lấy TheaterID từ URL parameter
	if theaterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing TheaterID"})
		return
	}

	type SeatData struct {
		SeatID      int    `json:"SeatID"`
		SeatNumber  int    `json:"SeatNumber"`
		Area        int    `json:"Area"`
		Column      int    `json:"Column"`
		Row         int    `json:"Row"`
		RowName     string `json:"RowName"`
		Description string `json:"Description"`
	}

	type RowData struct {
		Index int        `json:"index"`
		Name  string     `json:"name"`
		Seats []SeatData `json:"seats"`
	}

	var seatData []struct {
		SeatID      int    `json:"SeatID"`
		RowName     string `json:"RowName"`
		RowID       int    `json:"RowID"`
		SeatNumber  int    `json:"SeatNumber"`
		Area        int    `json:"Area"`
		Column      int    `json:"Column"`
		Row         int    `json:"Row"`
		Description string `json:"Description"`
	}

	// Sửa lại truy vấn để thay thế showtimeID bằng theaterID và lấy từ tham số URL
	err := database.DB.Raw(`
    SELECT *
    FROM seats
    WHERE RowID IN (
        SELECT RowID 
        FROM `+"`rows`"+`
        WHERE TheaterID = ?
    )
    ORDER BY RowID DESC, `+"`Column`"+` ASC
`, theaterID).Scan(&seatData).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	maxRow := 0
	maxColumn := 0
	rowIndexMap := make(map[string]int)
	rowNames := []string{}

	// Ghi nhận RowName theo thứ tự xuất hiện và tính maxRow
	for _, seat := range seatData {
		if _, exists := rowIndexMap[seat.RowName]; !exists {
			rowNames = append(rowNames, seat.RowName)
			rowIndexMap[seat.RowName] = len(rowNames) - 1 // Gán index theo thứ tự xuất hiện
		}
		if seat.Row > maxRow {
			maxRow = seat.Row
		}
		if seat.Column > maxColumn {
			maxColumn = seat.Column
		}
	}

	rowMap := make(map[string]*RowData)

	// Tạo RowData theo thứ tự đã ghi nhận
	for _, seat := range seatData {
		if _, exists := rowMap[seat.RowName]; !exists {
			rowMap[seat.RowName] = &RowData{
				Index: rowIndexMap[seat.RowName],
				Name:  seat.RowName,
				Seats: []SeatData{},
			}
		}
		rowMap[seat.RowName].Seats = append(rowMap[seat.RowName].Seats, SeatData{
			SeatID:      seat.SeatID,
			SeatNumber:  seat.SeatNumber,
			Area:        seat.Area,
			Column:      seat.Column,
			Row:         seat.Row,
			RowName:     seat.RowName,
			Description: seat.Description,
		})
	}

	// Chuyển map thành danh sách và sắp xếp theo Index
	rows := make([]RowData, 0, len(rowMap))
	for _, row := range rowMap {
		rows = append(rows, *row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Name == "A" { // Đưa hàng A xuống cuối
			return false
		}
		if rows[j].Name == "A" {
			return true
		}
		return rows[i].Name > rows[j].Name // Sắp xếp theo thứ tự giảm dần (Z → B)
	})

	result := gin.H{
		"maxColumn": maxColumn + 1, // Cộng thêm 1 để phản ánh index chính xác
		"maxRow":    maxRow,
		"rows":      rows,
	}

	c.JSON(http.StatusOK, result)
}

func AddTheater(c *gin.Context) {
	var theater models.Theater

	// Bind incoming JSON to the Branch struct
	if err := c.ShouldBindJSON(&theater); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	theater.CreatedAt = time.Now()
	theater.LastUpdatedAt = time.Now()

	if err := database.DisableForeignKeyChecks(database.DB, c); err != nil {
		return
	}

	// Save the new theater to the database
	if err := database.DB.Create(&theater).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create theater"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"data": theater})
}

func UpdateTheater(c *gin.Context) {
	id := c.Param("TheaterID") // Lấy TheaterID từ URL param

	var theater models.Theater
	// Kiểm tra xem rạp có tồn tại không
	if err := database.DB.First(&theater, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Theater not found"})
		return
	}

	// Parse dữ liệu JSON từ request body
	var updatedData models.Theater
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cập nhật các trường được cho phép
	theater.BranchID = updatedData.BranchID
	theater.TheaterName = updatedData.TheaterName
	theater.Slug = updatedData.Slug
	theater.TheaterType = updatedData.TheaterType
	theater.MaxRow = updatedData.MaxRow
	theater.MaxColumn = updatedData.MaxColumn
	theater.Status = updatedData.Status
	theater.LastUpdatedBy = updatedData.LastUpdatedBy
	theater.LastUpdatedAt = time.Now()

	// Cập nhật dữ liệu trong database
	if err := database.DB.Save(&theater).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theater"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Theater updated successfully",
		"data":    theater,
	})
}

func ChangeTheaterStatus(c *gin.Context) {
	// Lấy TheaterID từ param
	theaterID := c.Param("TheaterID")

	var theater models.Theater
	// Kiểm tra theater có tồn tại không
	if err := database.DB.First(&theater, "TheaterID = ?", theaterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Theater not found"})
		return
	}

	// Đảo trạng thái hiện tại
	theater.Status = !theater.Status
	theater.LastUpdatedAt = time.Now()

	// Lưu lại database
	if err := database.DB.Save(&theater).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theater status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"TheaterID": theater.TheaterID,
		"Status":    theater.Status,
	})
}
